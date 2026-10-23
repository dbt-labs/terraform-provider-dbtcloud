package analytics_credential

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = &analyticsCredentialResource{}
	_ resource.ResourceWithConfigure        = &analyticsCredentialResource{}
	_ resource.ResourceWithImportState      = &analyticsCredentialResource{}
	_ resource.ResourceWithModifyPlan       = &analyticsCredentialResource{}
	_ resource.ResourceWithConfigValidators = &analyticsCredentialResource{}
)

// AnalyticsCredentialResource returns the dbtcloud_analytics_credential resource
func AnalyticsCredentialResource() resource.Resource {
	return &analyticsCredentialResource{}
}

type analyticsCredentialResource struct {
	client *dbt_cloud.Client
}

func (r *analyticsCredentialResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_analytics_credential"
}

func (r *analyticsCredentialResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = analyticsCredentialSchema()
}

func (r *analyticsCredentialResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dbt_cloud.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *dbt_cloud.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	r.client = client
}

func (r *analyticsCredentialResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	var adapterBlocks []path.Expression
	for _, block := range supportedAdapterBlocks {
		adapterBlocks = append(adapterBlocks, path.MatchRoot(block))
	}

	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(adapterBlocks...),
		snowflakeAuthValidator{},
	}
}

// checkAdapter fails when the adapter block does not match the adapter of the connection
func (r *analyticsCredentialResource) checkAdapter(
	connectionID int64,
	block string,
	diags *diag.Diagnostics,
) {
	connection, err := r.client.GetGlobalConnectionAdapter(connectionID)
	if err != nil {
		diags.AddAttributeError(
			path.Root("connection_id"),
			"Error reading connection",
			fmt.Sprintf("Could not read connection %d to check its adapter: %s", connectionID, err),
		)
		return
	}

	expected, ok := adapterVersionToBlock[connection.Data.AdapterVersion]
	if !ok {
		diags.AddAttributeError(
			path.Root("connection_id"),
			"Unsupported adapter",
			fmt.Sprintf(
				"Connection %d uses the adapter %q, which does not support analytics credentials in this provider.",
				connectionID, connection.Data.AdapterVersion,
			),
		)
		return
	}

	if expected != block {
		diags.AddAttributeError(
			path.Root(block),
			"Adapter mismatch",
			fmt.Sprintf(
				"The `%s` block does not match connection %d, whose adapter is %q. Use the `%s` block, or point `connection_id` at a connection of the matching adapter.",
				block, connectionID, connection.Data.AdapterVersion, expected,
			),
		)
	}
}

func (r *analyticsCredentialResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	if req.Plan.Raw.IsNull() {
		// destroy plan
		return
	}

	var plan AnalyticsCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	block := plan.configuredBlock()
	if r.client != nil && block != "" && !plan.ConnectionID.IsUnknown() && !plan.ConnectionID.IsNull() {
		r.checkAdapter(plan.ConnectionID.ValueInt64(), block, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if req.State.Raw.IsNull() {
		return
	}

	var state AnalyticsCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// the adapter version comes from the connection, a new connection means a new one
	if !plan.ConnectionID.Equal(state.ConnectionID) {
		resp.Diagnostics.Append(
			resp.Plan.SetAttribute(ctx, path.Root("adapter_version"), types.StringUnknown())...,
		)
	}

	// an existing credential never changes adapter in place
	if (state.Snowflake == nil) != (plan.Snowflake == nil) {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root(blockSnowflake))
	}

	if state.Snowflake != nil && plan.Snowflake != nil {
		snowflakePath := path.Root(blockSnowflake)
		if !plan.Snowflake.AuthType.IsUnknown() && !plan.Snowflake.AuthType.Equal(state.Snowflake.AuthType) {
			resp.RequiresReplace = append(resp.RequiresReplace, snowflakePath.AtName("auth_type"))
		}
		if !state.Snowflake.Role.IsNull() && plan.Snowflake.Role.IsNull() {
			resp.Diagnostics.AddAttributeError(
				snowflakePath.AtName("role"),
				"Optional attribute cannot be removed",
				"The API cannot clear `role` once it is set. Set a new value, or recreate the credential "+
					"(for example with `terraform apply -replace=<resource address>`).",
			)
		}
	}
}

// snowflakeConfig builds the API config of the snowflake block. A secret is sent only on
// create or when its version changed; write-only values come from the configuration.
func snowflakeConfig(plan, config, state *SnowflakeModel) dbt_cloud.AnalyticsCredentialConfig {
	apiConfig := dbt_cloud.AnalyticsCredentialConfig{
		AuthType:  plan.AuthType.ValueString(),
		User:      plan.User.ValueString(),
		Role:      plan.Role.ValueString(),
		Warehouse: plan.Warehouse.ValueString(),
	}

	rotated := func(planVersion types.Int64, stateVersion func(*SnowflakeModel) types.Int64) bool {
		return state == nil || !planVersion.Equal(stateVersion(state))
	}

	if rotated(plan.PasswordWoVersion, func(s *SnowflakeModel) types.Int64 { return s.PasswordWoVersion }) {
		apiConfig.Password = helper.ResolveWriteOnlyString(config.PasswordWo, types.StringNull())
	}
	if rotated(plan.PrivateKeyWoVersion, func(s *SnowflakeModel) types.Int64 { return s.PrivateKeyWoVersion }) {
		apiConfig.PrivateKey = helper.ResolveWriteOnlyString(config.PrivateKeyWo, types.StringNull())
	}
	if rotated(plan.PrivateKeyPassphraseWoVersion, func(s *SnowflakeModel) types.Int64 { return s.PrivateKeyPassphraseWoVersion }) {
		apiConfig.PrivateKeyPassphrase = helper.ResolveWriteOnlyString(config.PrivateKeyPassphraseWo, types.StringNull())
	}

	return apiConfig
}

func (r *analyticsCredentialResource) apiConfig(
	plan, config, state *AnalyticsCredentialResourceModel,
) dbt_cloud.AnalyticsCredentialConfig {
	var stateSnowflake *SnowflakeModel
	if state != nil {
		stateSnowflake = state.Snowflake
	}
	return snowflakeConfig(plan.Snowflake, config.Snowflake, stateSnowflake)
}

func (r *analyticsCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan, config AnalyticsCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.checkAdapter(plan.ConnectionID.ValueInt64(), plan.configuredBlock(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateAnalyticsCredential(
		plan.ConnectionID.ValueInt64(),
		r.apiConfig(&plan, &config, nil),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating analytics credential", err.Error())
		return
	}
	if created.ID == nil {
		resp.Diagnostics.AddError(
			"Error creating analytics credential",
			"The API returned no ID for the created credential.",
		)
		return
	}

	// the state comes from the plan, never from the response
	plan.ID = types.Int64Value(*created.ID)
	plan.AdapterVersion = types.StringValue(created.AdapterVersion)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *analyticsCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state AnalyticsCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	credential, err := r.client.GetAnalyticsCredential(state.ID.ValueInt64())
	if err != nil {
		if helper.HandleResourceNotFound(ctx, err, &resp.Diagnostics, &resp.State, "analytics credential") {
			return
		}
		resp.Diagnostics.AddError("Error reading analytics credential", err.Error())
		return
	}

	state.ConnectionID = types.Int64Value(credential.ConnectionID)
	state.AdapterVersion = types.StringValue(credential.AdapterVersion)

	switch adapterVersionToBlock[credential.AdapterVersion] {
	case blockSnowflake:
		if state.Snowflake == nil {
			// import: nothing carried over, the versions stay null
			state.Snowflake = &SnowflakeModel{
				PasswordWo:             types.StringNull(),
				PasswordWoVersion:      types.Int64Null(),
				PrivateKeyWo:           types.StringNull(),
				PrivateKeyWoVersion:    types.Int64Null(),
				PrivateKeyPassphraseWo: types.StringNull(),
			}
			state.Snowflake.PrivateKeyPassphraseWoVersion = types.Int64Null()
		}
		// only the non-secret fields are read, the secrets are never in the state
		if credential.Config.AuthType != "" {
			state.Snowflake.AuthType = types.StringValue(credential.Config.AuthType)
		}
		state.Snowflake.User = types.StringValue(credential.Config.User)
		state.Snowflake.Warehouse = types.StringValue(credential.Config.Warehouse)
		state.Snowflake.Role = types.StringNull()
		if credential.Config.Role != "" {
			state.Snowflake.Role = types.StringValue(credential.Config.Role)
		}
	default:
		resp.Diagnostics.AddError(
			"Unsupported adapter",
			fmt.Sprintf(
				"The analytics credential %d uses the adapter %q, which this provider does not support.",
				state.ID.ValueInt64(), credential.AdapterVersion,
			),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *analyticsCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, config, state AnalyticsCredentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.checkAdapter(plan.ConnectionID.ValueInt64(), plan.configuredBlock(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateAnalyticsCredential(
		state.ID.ValueInt64(),
		r.apiConfig(&plan, &config, &state),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating analytics credential", err.Error())
		return
	}

	plan.ID = state.ID
	plan.AdapterVersion = types.StringValue(updated.AdapterVersion)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *analyticsCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state AnalyticsCredentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAnalyticsCredential(state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error deleting analytics credential", err.Error())
	}
}

// ImportState accepts "<credential_id>" and "<account_id>:<credential_id>"
func (r *analyticsCredentialResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.Split(req.ID, ":")
	if len(parts) > 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected `credential_id` or `account_id:credential_id`, got %q.", req.ID),
		)
		return
	}

	if len(parts) == 2 {
		accountID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || accountID != int64(r.client.AccountID) {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("The account of %q is not the account of the provider.", req.ID),
			)
			return
		}
	}

	credentialID, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Could not parse the credential ID of %q as an integer.", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), credentialID)...)
}
