package analytics_credential

import (
	"context"
	"fmt"
	"sort"
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
		bigqueryAuthValidator{},
	}
}

// checkAdapter fails when the adapter block does not match the adapter of the connection,
// or when the block uses an attribute the adapter version does not have; it returns the
// adapter version of the connection, empty when the connection could not be read
func (r *analyticsCredentialResource) checkAdapter(
	plan *AnalyticsCredentialResourceModel,
	diags *diag.Diagnostics,
) string {
	connectionID := plan.ConnectionID.ValueInt64()
	block := plan.configuredBlock()

	connection, err := r.client.GetGlobalConnectionAdapter(connectionID)
	if err != nil {
		diags.AddAttributeError(
			path.Root("connection_id"),
			"Error reading connection",
			fmt.Sprintf(
				"The provider reads the connection at plan time to check that the adapter block matches its adapter, so the token must be able to read connection %d. Could not read it: %s",
				connectionID, err,
			),
		)
		return ""
	}

	adapterVersion := connection.Data.AdapterVersion
	expected, ok := adapterVersionToBlock[adapterVersion]
	if !ok {
		diags.AddAttributeError(
			path.Root("connection_id"),
			"Unsupported adapter",
			fmt.Sprintf(
				"Connection %d uses the adapter %q, which does not support analytics credentials in this provider.",
				connectionID, adapterVersion,
			),
		)
		return ""
	}

	if expected != block {
		diags.AddAttributeError(
			path.Root(block),
			"Adapter mismatch",
			fmt.Sprintf(
				"The `%s` block does not match connection %d, whose adapter is %q. Use the `%s` block, or point `connection_id` at a connection of the matching adapter.",
				block, connectionID, adapterVersion, expected,
			),
		)
		return ""
	}

	// a bigquery_v0 credential is always a service account and has no auth type
	if adapterVersion == "bigquery_v0" && !plan.BigQuery.AuthType.IsNull() && !plan.BigQuery.AuthType.IsUnknown() {
		diags.AddAttributeError(
			path.Root(blockBigQuery).AtName("auth_type"),
			"Unsupported attribute for adapter",
			fmt.Sprintf(
				"Connection %d uses the adapter %q, which has no `auth_type`. Remove it, or use a connection of the latest BigQuery adapter.",
				connectionID, adapterVersion,
			),
		)
	}

	return adapterVersion
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
		r.checkAdapter(&plan, &resp.Diagnostics)
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
	statePresent := state.blocksPresent()
	for block, present := range plan.blocksPresent() {
		if present != statePresent[block] {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root(block))
		}
	}

	if state.Snowflake != nil && plan.Snowflake != nil {
		snowflakePath := path.Root(blockSnowflake)
		// an auth type that is not known yet may turn out different, and a replacement cannot be added later in the apply
		if plan.Snowflake.AuthType.IsUnknown() || !plan.Snowflake.AuthType.Equal(state.Snowflake.AuthType) {
			resp.RequiresReplace = append(resp.RequiresReplace, snowflakePath.AtName("auth_type"))
		}
	}
	if state.BigQuery != nil && plan.BigQuery != nil {
		bigqueryPath := path.Root(blockBigQuery)
		if plan.BigQuery.AuthType.IsUnknown() ||
			effectiveBigQueryAuthType(plan.BigQuery) != effectiveBigQueryAuthType(state.BigQuery) {
			resp.RequiresReplace = append(resp.RequiresReplace, bigqueryPath.AtName("auth_type"))
		}
	}

	// a replacement recreates the credential from scratch, so there is nothing to clear
	if len(resp.RequiresReplace) > 0 || !plan.ConnectionID.Equal(state.ConnectionID) {
		return
	}
	if state.Snowflake != nil && plan.Snowflake != nil {
		rejectRemoval(resp, state, path.Root(blockSnowflake).AtName("role"), state.Snowflake.Role, plan.Snowflake.Role)
		// switching the passphrase between its plain and write-only forms replaces the stored one, only dropping both clears it
		if plan.Snowflake.PrivateKeyPassphrase.IsNull() {
			rejectVersionRemoval(resp, state, path.Root(blockSnowflake).AtName("private_key_passphrase_wo_version"),
				state.Snowflake.PrivateKeyPassphraseWoVersion, plan.Snowflake.PrivateKeyPassphraseWoVersion)
		}
		if plan.Snowflake.PrivateKeyPassphraseWoVersion.IsNull() {
			rejectRemoval(resp, state, path.Root(blockSnowflake).AtName("private_key_passphrase"),
				state.Snowflake.PrivateKeyPassphrase, plan.Snowflake.PrivateKeyPassphrase)
		}
	}
	if state.BigQuery != nil && plan.BigQuery != nil {
		bigqueryPath := path.Root(blockBigQuery)
		rejectRemoval(resp, state, bigqueryPath.AtName("execution_project"), state.BigQuery.ExecutionProject, plan.BigQuery.ExecutionProject)
		rejectRemoval(resp, state, bigqueryPath.AtName("service_account_impersonation_url"),
			state.BigQuery.ServiceAccountImpersonationURL, plan.BigQuery.ServiceAccountImpersonationURL)
	}
	if state.Databricks != nil && plan.Databricks != nil {
		rejectRemoval(resp, state, path.Root(blockDatabricks).AtName("catalog"), state.Databricks.Catalog, plan.Databricks.Catalog)
	}
}

// effectiveBigQueryAuthType is the auth type the server uses: a missing one is a service
// account on the latest adapter, and stays unset on bigquery_v0
func effectiveBigQueryAuthType(bigquery *BigQueryModel) string {
	if bigquery.AuthType.IsNull() {
		return authTypeServiceAccount
	}
	return bigquery.AuthType.ValueString()
}

// rejectRemoval fails the plan when an optional attribute that is set in the state is removed:
// the API cannot clear it
func rejectRemoval(resp *resource.ModifyPlanResponse, credential AnalyticsCredentialResourceModel, attribute path.Path, stateValue, planValue types.String) {
	if !stateValue.IsNull() && planValue.IsNull() {
		resp.Diagnostics.AddAttributeError(
			attribute,
			"Optional attribute cannot be removed",
			fmt.Sprintf(
				"The API cannot clear `%s` of credential %d (connection %d) once it is set. Set a new value, or recreate the credential: destroy it with its current configuration "+
					"(`terraform destroy -target=<resource address>`), then apply with the attribute removed.",
				attribute.String(), credential.ID.ValueInt64(), credential.ConnectionID.ValueInt64(),
			),
		)
	}
}

// rejectVersionRemoval fails the plan when the version of an optional secret is removed: the
// API cannot clear the secret, so the state would stop tracking a value the server still has
func rejectVersionRemoval(resp *resource.ModifyPlanResponse, credential AnalyticsCredentialResourceModel, attribute path.Path, stateValue, planValue types.Int64) {
	if !stateValue.IsNull() && planValue.IsNull() {
		resp.Diagnostics.AddAttributeError(
			attribute,
			"Optional attribute cannot be removed",
			fmt.Sprintf(
				"The API cannot clear the secret behind `%s` of credential %d (connection %d) once it is set. Keep the secret and its version, or recreate the credential: destroy it with its current configuration "+
					"(`terraform destroy -target=<resource address>`), then apply with the attribute removed.",
				attribute.String(), credential.ID.ValueInt64(), credential.ConnectionID.ValueInt64(),
			),
		)
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

	hasState := state != nil
	var previous SnowflakeModel
	if hasState {
		previous = *state
	}
	apiConfig.Password = secretValue(
		plan.Password, previous.Password, config.PasswordWo, plan.PasswordWoVersion, previous.PasswordWoVersion, hasState,
	)
	apiConfig.PrivateKey = secretValue(
		plan.PrivateKey, previous.PrivateKey, config.PrivateKeyWo, plan.PrivateKeyWoVersion, previous.PrivateKeyWoVersion, hasState,
	)
	apiConfig.PrivateKeyPassphrase = secretValue(
		plan.PrivateKeyPassphrase, previous.PrivateKeyPassphrase,
		config.PrivateKeyPassphraseWo, plan.PrivateKeyPassphraseWoVersion, previous.PrivateKeyPassphraseWoVersion, hasState,
	)

	return apiConfig
}

func (r *analyticsCredentialResource) apiConfig(
	plan, config, state *AnalyticsCredentialResourceModel,
) dbt_cloud.AnalyticsCredentialConfig {
	if plan.BigQuery != nil {
		var stateBigQuery *BigQueryModel
		if state != nil {
			stateBigQuery = state.BigQuery
		}
		return bigqueryConfig(plan.BigQuery, config.BigQuery, stateBigQuery)
	}

	if plan.Redshift != nil {
		var stateRedshift *RedshiftModel
		if state != nil {
			stateRedshift = state.Redshift
		}
		return redshiftConfig(plan.Redshift, config.Redshift, stateRedshift)
	}

	if plan.Databricks != nil {
		var stateDatabricks *DatabricksModel
		if state != nil {
			stateDatabricks = state.Databricks
		}
		return databricksConfig(plan.Databricks, config.Databricks, stateDatabricks)
	}

	var stateSnowflake *SnowflakeModel
	if state != nil {
		stateSnowflake = state.Snowflake
	}
	return snowflakeConfig(plan.Snowflake, config.Snowflake, stateSnowflake)
}

// bigqueryConfig builds the API config of the bigquery block; the private key is sent only
// on create or when its version changed
func bigqueryConfig(plan, config, state *BigQueryModel) dbt_cloud.AnalyticsCredentialConfig {
	apiConfig := dbt_cloud.AnalyticsCredentialConfig{
		AuthType:                       plan.AuthType.ValueString(),
		ExecutionProject:               plan.ExecutionProject.ValueString(),
		PrivateKeyID:                   plan.PrivateKeyID.ValueString(),
		ClientEmail:                    plan.ClientEmail.ValueString(),
		ClientID:                       plan.ClientID.ValueString(),
		AuthURI:                        plan.AuthURI.ValueString(),
		TokenURI:                       plan.TokenURI.ValueString(),
		AuthProviderX509CertURL:        plan.AuthProviderX509CertURL.ValueString(),
		ClientX509CertURL:              plan.ClientX509CertURL.ValueString(),
		WorkloadPoolProviderPath:       plan.WorkloadPoolProviderPath.ValueString(),
		ServiceAccountImpersonationURL: plan.ServiceAccountImpersonationURL.ValueString(),
	}
	var previous BigQueryModel
	if state != nil {
		previous = *state
	}
	apiConfig.PrivateKey = secretValue(
		plan.PrivateKey, previous.PrivateKey, config.PrivateKeyWo, plan.PrivateKeyWoVersion, previous.PrivateKeyWoVersion, state != nil,
	)
	return apiConfig
}

// stringOrNull maps the empty strings of the API to null
func stringOrNull(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

// redshiftConfig builds the API config of the redshift block; the password is sent only
// on create or when its version changed
func redshiftConfig(plan, config, state *RedshiftModel) dbt_cloud.AnalyticsCredentialConfig {
	apiConfig := dbt_cloud.AnalyticsCredentialConfig{Username: plan.Username.ValueString()}
	var previous RedshiftModel
	if state != nil {
		previous = *state
	}
	apiConfig.Password = secretValue(
		plan.Password, previous.Password, config.PasswordWo, plan.PasswordWoVersion, previous.PasswordWoVersion, state != nil,
	)
	return apiConfig
}

// databricksConfig builds the API config of the databricks block; the token is sent only
// on create or when its version changed
func databricksConfig(plan, config, state *DatabricksModel) dbt_cloud.AnalyticsCredentialConfig {
	apiConfig := dbt_cloud.AnalyticsCredentialConfig{Catalog: plan.Catalog.ValueString()}
	var previous DatabricksModel
	if state != nil {
		previous = *state
	}
	apiConfig.Token = secretValue(
		plan.Token, previous.Token, config.TokenWo, plan.TokenWoVersion, previous.TokenWoVersion, state != nil,
	)
	return apiConfig
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

	adapterVersion := r.checkAdapter(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	apiConfig := r.apiConfig(&plan, &config, nil)
	// the latest BigQuery adapter has no server side default: a missing auth type is a service account
	if plan.BigQuery != nil && adapterVersion != "bigquery_v0" && apiConfig.AuthType == "" {
		apiConfig.AuthType = authTypeServiceAccount
	}

	created, err := r.client.CreateAnalyticsCredential(
		plan.ConnectionID.ValueInt64(),
		apiConfig,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating analytics credential",
			fmt.Sprintf("Could not create the analytics credential on connection %d: %s", plan.ConnectionID.ValueInt64(), err),
		)
		return
	}
	if created.ID == nil {
		resp.Diagnostics.AddError(
			"Error creating analytics credential",
			fmt.Sprintf(
				"The API accepted the request for connection %d but returned no ID, so a credential may already exist on it. Check the connection in dbt and, if there is one, import it with `terraform import` and its credential ID.",
				plan.ConnectionID.ValueInt64(),
			),
		)
		return
	}

	// the state comes from the plan, never from the response
	plan.ID = types.Int64Value(*created.ID)
	plan.AdapterVersion = reportedAdapterVersion(types.StringValue(adapterVersion), created.AdapterVersion)
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
		resp.Diagnostics.AddError(
			"Error reading analytics credential",
			fmt.Sprintf("Could not read analytics credential %d: %s", state.ID.ValueInt64(), err),
		)
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
		state.Snowflake.Role = stringOrNull(credential.Config.Role)
	case blockDatabricks:
		if state.Databricks == nil {
			// import: nothing carried over, the version stays null
			state.Databricks = &DatabricksModel{TokenWo: types.StringNull(), TokenWoVersion: types.Int64Null()}
		}
		state.Databricks.Catalog = stringOrNull(credential.Config.Catalog)
	case blockRedshift:
		if state.Redshift == nil {
			// import: nothing carried over, the version stays null
			state.Redshift = &RedshiftModel{PasswordWo: types.StringNull(), PasswordWoVersion: types.Int64Null()}
		}
		state.Redshift.Username = types.StringValue(credential.Config.Username)
	case blockBigQuery:
		imported := state.BigQuery == nil
		if imported {
			// import: nothing carried over, the version stays null
			state.BigQuery = &BigQueryModel{
				AuthType:            types.StringNull(),
				PrivateKeyWo:        types.StringNull(),
				PrivateKeyWoVersion: types.Int64Null(),
			}
		}
		bigquery := state.BigQuery
		// the latest adapter defaults to a service account: keep an auth_type that was left out of the configuration
		// unset, but after an import there is no configuration to follow, so store what the server returns
		apiAuthType := credential.Config.AuthType
		if apiAuthType != "" && !(apiAuthType == authTypeServiceAccount && bigquery.AuthType.IsNull() && !imported) {
			bigquery.AuthType = types.StringValue(apiAuthType)
		}
		bigquery.ExecutionProject = stringOrNull(credential.Config.ExecutionProject)
		bigquery.PrivateKeyID = stringOrNull(credential.Config.PrivateKeyID)
		bigquery.ClientEmail = stringOrNull(credential.Config.ClientEmail)
		bigquery.ClientID = stringOrNull(credential.Config.ClientID)
		bigquery.AuthURI = stringOrNull(credential.Config.AuthURI)
		bigquery.TokenURI = stringOrNull(credential.Config.TokenURI)
		bigquery.AuthProviderX509CertURL = stringOrNull(credential.Config.AuthProviderX509CertURL)
		bigquery.ClientX509CertURL = stringOrNull(credential.Config.ClientX509CertURL)
		bigquery.WorkloadPoolProviderPath = stringOrNull(credential.Config.WorkloadPoolProviderPath)
		bigquery.ServiceAccountImpersonationURL = stringOrNull(credential.Config.ServiceAccountImpersonationURL)
	default:
		resp.Diagnostics.AddError(
			"Unsupported adapter",
			fmt.Sprintf(
				"The analytics credential %d uses the adapter %q, which this provider does not support (supported: %s). Remove it from the state with `terraform state rm`, or manage a credential of a supported adapter.",
				state.ID.ValueInt64(), credential.AdapterVersion, supportedAdapterVersions(),
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

	r.checkAdapter(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateAnalyticsCredential(
		state.ID.ValueInt64(),
		r.apiConfig(&plan, &config, &state),
	)
	if err != nil {
		detail := fmt.Sprintf(
			"Could not update analytics credential %d (connection %d): %s",
			state.ID.ValueInt64(), state.ConnectionID.ValueInt64(), err,
		)
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			detail += "\n\nThe credential no longer exists, or the token cannot see it. Run `terraform plan` to refresh the state, or recreate it with `terraform apply -replace=<address>`."
		}
		resp.Diagnostics.AddError("Error updating analytics credential", detail)
		return
	}

	plan.ID = state.ID
	plan.AdapterVersion = reportedAdapterVersion(state.AdapterVersion, updated.AdapterVersion)
	applyReportedConfig(&plan, updated)
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
		// already gone: nothing left to delete
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting analytics credential",
			fmt.Sprintf(
				"Could not delete analytics credential %d (connection %d): %s",
				state.ID.ValueInt64(), state.ConnectionID.ValueInt64(), err,
			),
		)
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
				fmt.Sprintf("The account of %q is not the account of the provider, which is %d.", req.ID, r.client.AccountID),
			)
			return
		}
	}

	credentialID, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Could not parse the credential ID of %q as an integer: %s", req.ID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), credentialID)...)
}

// supportedAdapterVersions lists the adapter versions with an analytics credential block,
// for error messages
func supportedAdapterVersions() string {
	versions := make([]string, 0, len(adapterVersionToBlock))
	for version := range adapterVersionToBlock {
		versions = append(versions, version)
	}
	sort.Strings(versions)
	return strings.Join(versions, ", ")
}

// secretValue returns the secret to put in a request. A plain secret is sent on create and when
// it changed; a write-only one follows secretToSend.
func secretValue(
	plain, previousPlain, writeOnly types.String,
	planVersion, previousVersion types.Int64,
	hasState bool,
) string {
	if !plain.IsNull() && !plain.IsUnknown() {
		if hasState && plain.Equal(previousPlain) {
			return ""
		}
		return plain.ValueString()
	}
	return secretToSend(writeOnly, planVersion, previousVersion, hasState)
}

// secretToSend returns the write-only secret to put in a request. A secret is sent on create
// and when its version changed, and left out otherwise so that the server keeps the stored
// one. Without a version nothing is sent: a secret must come with a version.
func secretToSend(secret types.String, planVersion, stateVersion types.Int64, hasState bool) string {
	if planVersion.IsNull() {
		return ""
	}
	if hasState && planVersion.Equal(stateVersion) {
		return ""
	}
	return helper.ResolveWriteOnlyString(secret, types.StringNull())
}

// reportedAdapterVersion returns the adapter version the API reports, or the fallback when the
// response leaves it out.
func reportedAdapterVersion(fallback types.String, reported string) types.String {
	if reported != "" {
		return types.StringValue(reported)
	}
	if fallback.IsUnknown() || fallback.ValueString() == "" {
		return types.StringNull()
	}
	return fallback
}

// reportedString returns the value the API reports for a non-secret field. A field left out of the
// configuration stays null, because the server may fill in a default, and a field the response
// leaves empty keeps the planned value.
func reportedString(planned types.String, reported string) types.String {
	if planned.IsNull() || planned.IsUnknown() || reported == "" {
		return planned
	}
	return types.StringValue(reported)
}

// applyReportedConfig sets the non-secret fields of the configured adapter block from the API
// response. Secrets are never taken from it.
func applyReportedConfig(model *AnalyticsCredentialResourceModel, credential *dbt_cloud.AnalyticsCredential) {
	config := credential.Config
	switch {
	case model.Snowflake != nil:
		block := model.Snowflake
		block.AuthType = reportedString(block.AuthType, config.AuthType)
		block.User = reportedString(block.User, config.User)
		block.Role = reportedString(block.Role, config.Role)
		block.Warehouse = reportedString(block.Warehouse, config.Warehouse)
	case model.Databricks != nil:
		model.Databricks.Catalog = reportedString(model.Databricks.Catalog, config.Catalog)
	case model.Redshift != nil:
		model.Redshift.Username = reportedString(model.Redshift.Username, config.Username)
	case model.BigQuery != nil:
		block := model.BigQuery
		block.AuthType = reportedString(block.AuthType, config.AuthType)
		block.ExecutionProject = reportedString(block.ExecutionProject, config.ExecutionProject)
		block.PrivateKeyID = reportedString(block.PrivateKeyID, config.PrivateKeyID)
		block.ClientEmail = reportedString(block.ClientEmail, config.ClientEmail)
		block.ClientID = reportedString(block.ClientID, config.ClientID)
		block.AuthURI = reportedString(block.AuthURI, config.AuthURI)
		block.TokenURI = reportedString(block.TokenURI, config.TokenURI)
		block.AuthProviderX509CertURL = reportedString(block.AuthProviderX509CertURL, config.AuthProviderX509CertURL)
		block.ClientX509CertURL = reportedString(block.ClientX509CertURL, config.ClientX509CertURL)
		block.WorkloadPoolProviderPath = reportedString(block.WorkloadPoolProviderPath, config.WorkloadPoolProviderPath)
		block.ServiceAccountImpersonationURL = reportedString(block.ServiceAccountImpersonationURL, config.ServiceAccountImpersonationURL)
	}
}
