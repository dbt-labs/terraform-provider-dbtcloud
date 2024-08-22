package global_connection

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

var (
	_ resource.Resource                     = &globalConnectionResource{}
	_ resource.ResourceWithConfigure        = &globalConnectionResource{}
	_ resource.ResourceWithImportState      = &globalConnectionResource{}
	_ resource.ResourceWithConfigValidators = &globalConnectionResource{}
	_ resource.ResourceWithModifyPlan       = &globalConnectionResource{}
)

func GlobalConnectionResource() resource.Resource {
	return &globalConnectionResource{}
}

type globalConnectionResource struct {
	client *dbt_cloud.Client
}

func (r *globalConnectionResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_global_connection"
}

func (r globalConnectionResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("snowflake"),
			path.MatchRoot("bigquery"),
		),
	}
}

func (r globalConnectionResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {

	var plan, state GlobalConnectionResourceModel

	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		// we only check when both plan and state are not null
		return
	}

	// Read the current state and planned state
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	type ConfigState struct {
		WasNull bool
		IsNull  bool
	}

	configStates := map[string]ConfigState{
		"bigquery": {
			WasNull: state.BigQueryConfig == nil,
			IsNull:  plan.BigQueryConfig == nil,
		},
		"snowflake": {
			WasNull: state.SnowflakeConfig == nil,
			IsNull:  plan.SnowflakeConfig == nil,
		},
		// Add more types here as needed
	}

	configStatesVals := lo.Keys(configStates)
	left, right := lo.Difference(configStatesVals, supportedGlobalConfigTypes)
	if len(left) > 0 || len(right) > 0 {
		panic(
			"ModifyPlan is missing some of the Data Warehouse types. The provider needs to be updated",
		)
	}

	for configType, configState := range configStates {
		if (configState.WasNull && !configState.IsNull) ||
			(!configState.WasNull && configState.IsNull) {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root(configType))
		}
	}

}

func (r *globalConnectionResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state GlobalConnectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	connectionID := state.ID.ValueInt64()

	switch {
	case state.SnowflakeConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.SnowflakeConfig](r.client)

		common, snowflakeCfg, err := c.Get(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				resp.Diagnostics.AddWarning(
					"Resource not found",
					"The connection resource was not found and has been removed from the state.",
				)
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Error getting the connection", err.Error())
			return
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(snowflakeCfg.AdapterVersion())
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)
		state.OauthConfigurationId = types.Int64PointerValue(common.OauthConfigurationId)

		// nullable common fields
		if !common.PrivateLinkEndpointId.IsNull() {
			state.PrivateLinkEndpointId = types.StringValue(common.PrivateLinkEndpointId.MustGet())
		} else {
			state.PrivateLinkEndpointId = types.StringNull()
		}

		// snowflake settings
		state.SnowflakeConfig.Account = types.StringPointerValue(snowflakeCfg.Account)
		state.SnowflakeConfig.Database = types.StringPointerValue(snowflakeCfg.Database)
		state.SnowflakeConfig.Warehouse = types.StringPointerValue(snowflakeCfg.Warehouse)
		state.SnowflakeConfig.ClientSessionKeepAlive = types.BoolPointerValue(
			snowflakeCfg.ClientSessionKeepAlive,
		)
		state.SnowflakeConfig.AllowSso = types.BoolPointerValue(snowflakeCfg.AllowSso)

		// nullable optional fields
		// TODO: decide if it is better to read it as string, *string or nullable.Nullable[string] on the dbt_cloud side
		// in this case role can never be empty so this works but we might have cases where null and empty are different
		if !snowflakeCfg.Role.IsNull() {
			state.SnowflakeConfig.Role = types.StringValue(snowflakeCfg.Role.MustGet())
		} else {
			state.SnowflakeConfig.Role = types.StringNull()
		}

		// We don't set the sensitive fields when we read because those are secret and never returned by the API
		// sensitive fields: OauthClientID, OauthClientSecret

	case state.BigQueryConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.BigQueryConfig](r.client)

		common, bigqueryCfg, err := c.Get(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				resp.Diagnostics.AddWarning(
					"Resource not found",
					"The connection resource was not found and has been removed from the state.",
				)
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Error getting the connection", err.Error())
			return
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(bigqueryCfg.AdapterVersion())
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)
		state.OauthConfigurationId = types.Int64PointerValue(common.OauthConfigurationId)

		// nullable common fields
		if !common.PrivateLinkEndpointId.IsNull() {
			state.PrivateLinkEndpointId = types.StringValue(common.PrivateLinkEndpointId.MustGet())
		} else {
			state.PrivateLinkEndpointId = types.StringNull()
		}

		// BigQuery settings
		state.BigQueryConfig.GCPProjectID = types.StringPointerValue(bigqueryCfg.ProjectID)
		state.BigQueryConfig.TimeoutSeconds = types.Int64PointerValue(bigqueryCfg.TimeoutSeconds)
		state.BigQueryConfig.PrivateKeyID = types.StringPointerValue(bigqueryCfg.PrivateKeyID)
		state.BigQueryConfig.ClientEmail = types.StringPointerValue(bigqueryCfg.ClientEmail)
		state.BigQueryConfig.ClientID = types.StringPointerValue(bigqueryCfg.ClientID)
		state.BigQueryConfig.AuthURI = types.StringPointerValue(bigqueryCfg.AuthURI)
		state.BigQueryConfig.TokenURI = types.StringPointerValue(bigqueryCfg.TokenURI)
		state.BigQueryConfig.AuthProviderX509CertURL = types.StringPointerValue(
			bigqueryCfg.AuthProviderX509CertURL,
		)
		state.BigQueryConfig.ClientX509CertURL = types.StringPointerValue(
			bigqueryCfg.ClientX509CertURL,
		)
		state.BigQueryConfig.Retries = types.Int64PointerValue(bigqueryCfg.Retries)
		state.BigQueryConfig.Scopes = helper.SliceStringToSliceTypesString(bigqueryCfg.Scopes)

		// nullable optional fields
		if !bigqueryCfg.Priority.IsNull() {
			state.BigQueryConfig.Priority = types.StringValue(bigqueryCfg.Priority.MustGet())
		} else {
			state.BigQueryConfig.Priority = types.StringNull()
		}

		if !bigqueryCfg.Location.IsNull() {
			state.BigQueryConfig.Location = types.StringValue(bigqueryCfg.Location.MustGet())
		} else {
			state.BigQueryConfig.Location = types.StringNull()
		}

		if !bigqueryCfg.MaximumBytesBilled.IsNull() {
			state.BigQueryConfig.MaximumBytesBilled = types.Int64Value(
				bigqueryCfg.MaximumBytesBilled.MustGet(),
			)
		} else {
			state.BigQueryConfig.MaximumBytesBilled = types.Int64Null()
		}

		if !bigqueryCfg.ExecutionProject.IsNull() {
			state.BigQueryConfig.ExecutionProject = types.StringValue(
				bigqueryCfg.ExecutionProject.MustGet(),
			)
		} else {
			state.BigQueryConfig.ExecutionProject = types.StringNull()
		}

		if !bigqueryCfg.ImpersonateServiceAccount.IsNull() {
			state.BigQueryConfig.ImpersonateServiceAccount = types.StringValue(
				bigqueryCfg.ImpersonateServiceAccount.MustGet(),
			)
		} else {
			state.BigQueryConfig.ImpersonateServiceAccount = types.StringNull()
		}

		if !bigqueryCfg.JobRetryDeadlineSeconds.IsNull() {
			state.BigQueryConfig.JobRetryDeadlineSeconds = types.Int64Value(
				bigqueryCfg.JobRetryDeadlineSeconds.MustGet(),
			)
		} else {
			state.BigQueryConfig.JobRetryDeadlineSeconds = types.Int64Null()
		}

		if !bigqueryCfg.JobCreationTimeoutSeconds.IsNull() {
			state.BigQueryConfig.JobCreationTimeoutSeconds = types.Int64Value(
				bigqueryCfg.JobCreationTimeoutSeconds.MustGet(),
			)
		} else {
			state.BigQueryConfig.JobCreationTimeoutSeconds = types.Int64Null()
		}

		if !bigqueryCfg.GcsBucket.IsNull() {
			state.BigQueryConfig.GcsBucket = types.StringValue(bigqueryCfg.GcsBucket.MustGet())
		} else {
			state.BigQueryConfig.GcsBucket = types.StringNull()
		}

		if !bigqueryCfg.DataprocRegion.IsNull() {
			state.BigQueryConfig.DataprocRegion = types.StringValue(
				bigqueryCfg.DataprocRegion.MustGet(),
			)
		} else {
			state.BigQueryConfig.DataprocRegion = types.StringNull()
		}

		if !bigqueryCfg.DataprocClusterName.IsNull() {
			state.BigQueryConfig.DataprocClusterName = types.StringValue(
				bigqueryCfg.DataprocClusterName.MustGet(),
			)
		} else {
			state.BigQueryConfig.DataprocClusterName = types.StringNull()
		}

		// We don't set the sensitive fields when we read because those are secret and never returned by the API
		// sensitive fields: ApplicationID, ApplicationSecret, PrivateKey

	default:
		panic("Unknown connection type")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *globalConnectionResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan GlobalConnectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	commonCfg := dbt_cloud.GlobalConnectionCommon{
		Name: plan.Name.ValueStringPointer(),
	}

	// nullable common fields
	if !plan.PrivateLinkEndpointId.IsNull() {
		commonCfg.PrivateLinkEndpointId.Set(plan.PrivateLinkEndpointId.ValueString())
	}

	// data warehouse specific
	switch {
	case plan.SnowflakeConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.SnowflakeConfig](r.client)

		snowflakeCfg := dbt_cloud.SnowflakeConfig{
			Account:                plan.SnowflakeConfig.Account.ValueStringPointer(),
			Database:               plan.SnowflakeConfig.Database.ValueStringPointer(),
			Warehouse:              plan.SnowflakeConfig.Warehouse.ValueStringPointer(),
			ClientSessionKeepAlive: plan.SnowflakeConfig.ClientSessionKeepAlive.ValueBoolPointer(),
			AllowSso:               plan.SnowflakeConfig.AllowSso.ValueBoolPointer(),
			OauthClientID:          plan.SnowflakeConfig.OauthClientID.ValueStringPointer(),
			OauthClientSecret:      plan.SnowflakeConfig.OauthClientSecret.ValueStringPointer(),
		}

		// nullable fields
		if !plan.SnowflakeConfig.Role.IsNull() {
			snowflakeCfg.Role.Set(plan.SnowflakeConfig.Role.ValueString())
		}

		commonResp, _, err := c.Create(commonCfg, snowflakeCfg)

		if err != nil {
			resp.Diagnostics.AddError("Error creating the connection", err.Error())
			return
		}

		// we set the computed values that don't have any default
		plan.ID = types.Int64PointerValue(commonResp.ID)
		plan.AdapterVersion = types.StringValue(snowflakeCfg.AdapterVersion())
		plan.OauthConfigurationId = types.Int64PointerValue(commonResp.OauthConfigurationId)
		plan.IsSshTunnelEnabled = types.BoolPointerValue(commonResp.IsSshTunnelEnabled)

	case plan.BigQueryConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.BigQueryConfig](r.client)

		bigqueryCfg := dbt_cloud.BigQueryConfig{
			ProjectID:               plan.BigQueryConfig.GCPProjectID.ValueStringPointer(),
			TimeoutSeconds:          plan.BigQueryConfig.TimeoutSeconds.ValueInt64Pointer(),
			PrivateKeyID:            plan.BigQueryConfig.PrivateKeyID.ValueStringPointer(),
			PrivateKey:              plan.BigQueryConfig.PrivateKey.ValueStringPointer(),
			ClientEmail:             plan.BigQueryConfig.ClientEmail.ValueStringPointer(),
			ClientID:                plan.BigQueryConfig.ClientID.ValueStringPointer(),
			AuthURI:                 plan.BigQueryConfig.AuthURI.ValueStringPointer(),
			TokenURI:                plan.BigQueryConfig.TokenURI.ValueStringPointer(),
			AuthProviderX509CertURL: plan.BigQueryConfig.AuthProviderX509CertURL.ValueStringPointer(),
			ClientX509CertURL:       plan.BigQueryConfig.ClientX509CertURL.ValueStringPointer(),
			Retries:                 plan.BigQueryConfig.Retries.ValueInt64Pointer(),
			Scopes: helper.TypesStringSliceToStringSlice(
				plan.BigQueryConfig.Scopes,
			),
		}

		// nullable fields
		if !plan.BigQueryConfig.Priority.IsNull() {
			bigqueryCfg.Priority.Set(plan.BigQueryConfig.Priority.ValueString())
		}
		if !plan.BigQueryConfig.Location.IsNull() {
			bigqueryCfg.Location.Set(plan.BigQueryConfig.Location.ValueString())
		}
		if !plan.BigQueryConfig.MaximumBytesBilled.IsNull() {
			bigqueryCfg.MaximumBytesBilled.Set(plan.BigQueryConfig.MaximumBytesBilled.ValueInt64())
		}
		if !plan.BigQueryConfig.ExecutionProject.IsNull() {
			bigqueryCfg.ExecutionProject.Set(plan.BigQueryConfig.ExecutionProject.ValueString())
		}
		if !plan.BigQueryConfig.ImpersonateServiceAccount.IsNull() {
			bigqueryCfg.ImpersonateServiceAccount.Set(
				plan.BigQueryConfig.ImpersonateServiceAccount.ValueString(),
			)
		}
		if !plan.BigQueryConfig.JobRetryDeadlineSeconds.IsNull() {
			bigqueryCfg.JobRetryDeadlineSeconds.Set(
				plan.BigQueryConfig.JobRetryDeadlineSeconds.ValueInt64(),
			)
		}
		if !plan.BigQueryConfig.JobCreationTimeoutSeconds.IsNull() {
			bigqueryCfg.JobCreationTimeoutSeconds.Set(
				plan.BigQueryConfig.JobCreationTimeoutSeconds.ValueInt64(),
			)
		}
		if !plan.BigQueryConfig.ApplicationID.IsNull() {
			bigqueryCfg.ApplicationID.Set(plan.BigQueryConfig.ApplicationID.ValueString())
		}
		if !plan.BigQueryConfig.ApplicationSecret.IsNull() {
			bigqueryCfg.ApplicationSecret.Set(plan.BigQueryConfig.ApplicationSecret.ValueString())
		}
		if !plan.BigQueryConfig.GcsBucket.IsNull() {
			bigqueryCfg.GcsBucket.Set(plan.BigQueryConfig.GcsBucket.ValueString())
		}
		if !plan.BigQueryConfig.DataprocRegion.IsNull() {
			bigqueryCfg.DataprocRegion.Set(plan.BigQueryConfig.DataprocRegion.ValueString())
		}
		if !plan.BigQueryConfig.DataprocClusterName.IsNull() {
			bigqueryCfg.DataprocClusterName.Set(
				plan.BigQueryConfig.DataprocClusterName.ValueString(),
			)
		}

		commonResp, _, err := c.Create(commonCfg, bigqueryCfg)

		if err != nil {
			resp.Diagnostics.AddError("Error creating the connection", err.Error())
			return
		}

		// we set the computed values that don't have any default
		plan.ID = types.Int64PointerValue(commonResp.ID)
		plan.AdapterVersion = types.StringValue(bigqueryCfg.AdapterVersion())
		plan.OauthConfigurationId = types.Int64PointerValue(commonResp.OauthConfigurationId)
		plan.IsSshTunnelEnabled = types.BoolPointerValue(commonResp.IsSshTunnelEnabled)

	default:
		panic("Unknown connection type")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *globalConnectionResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state GlobalConnectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionID := state.ID.ValueInt64()

	_, err := r.client.DeleteGlobalConnection(connectionID)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting the connection", err.Error())
		return
	}

}

func (r *globalConnectionResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state GlobalConnectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	globalConfigChanges := dbt_cloud.GlobalConnectionCommon{}

	if plan.Name != state.Name {
		globalConfigChanges.Name = plan.Name.ValueStringPointer()
	}
	// nullable common fields
	if plan.PrivateLinkEndpointId != state.PrivateLinkEndpointId {
		if plan.PrivateLinkEndpointId.IsNull() {
			globalConfigChanges.PrivateLinkEndpointId.SetNull()
		} else {
			globalConfigChanges.PrivateLinkEndpointId.Set(plan.PrivateLinkEndpointId.ValueString())
		}
	}

	switch {
	case plan.SnowflakeConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.SnowflakeConfig](r.client)

		warehouseConfigChanges := dbt_cloud.SnowflakeConfig{}

		// Snowflake specific ones
		if plan.SnowflakeConfig.Account != state.SnowflakeConfig.Account {
			warehouseConfigChanges.Account = plan.SnowflakeConfig.Account.ValueStringPointer()
		}
		if plan.SnowflakeConfig.Database != state.SnowflakeConfig.Database {
			warehouseConfigChanges.Database = plan.SnowflakeConfig.Database.ValueStringPointer()
		}
		if plan.SnowflakeConfig.Warehouse != state.SnowflakeConfig.Warehouse {
			warehouseConfigChanges.Warehouse = plan.SnowflakeConfig.Warehouse.ValueStringPointer()
		}
		if plan.SnowflakeConfig.ClientSessionKeepAlive != state.SnowflakeConfig.ClientSessionKeepAlive {
			warehouseConfigChanges.ClientSessionKeepAlive = plan.SnowflakeConfig.ClientSessionKeepAlive.ValueBoolPointer()
		}
		if plan.SnowflakeConfig.AllowSso != state.SnowflakeConfig.AllowSso {
			warehouseConfigChanges.AllowSso = plan.SnowflakeConfig.AllowSso.ValueBoolPointer()
		}
		if plan.SnowflakeConfig.OauthClientID != state.SnowflakeConfig.OauthClientID {
			warehouseConfigChanges.OauthClientID = plan.SnowflakeConfig.OauthClientID.ValueStringPointer()
		}
		if plan.SnowflakeConfig.OauthClientSecret != state.SnowflakeConfig.OauthClientSecret {
			warehouseConfigChanges.OauthClientSecret = plan.SnowflakeConfig.OauthClientSecret.ValueStringPointer()
		}

		// nullable fields
		// when the values are Null, we still want to send it as null to the PATCH payload, to remove it, otherwise the omitempty doesn't add it to the payload and it doesn't get updated
		if plan.SnowflakeConfig.Role != state.SnowflakeConfig.Role {
			if plan.SnowflakeConfig.Role.IsNull() {
				warehouseConfigChanges.Role.SetNull()
			} else {
				warehouseConfigChanges.Role.Set(plan.SnowflakeConfig.Role.ValueString())
			}
		}

		updateCommon, _, err := c.Update(
			state.ID.ValueInt64(),
			globalConfigChanges,
			warehouseConfigChanges,
		)
		if err != nil {
			resp.Diagnostics.AddError("Error updating global connection", err.Error())
			return
		}

		// we set the computed values, no need to do it for ID as we use a PlanModifier with UseStateForUnknown()
		plan.IsSshTunnelEnabled = types.BoolPointerValue(updateCommon.IsSshTunnelEnabled)
		plan.OauthConfigurationId = types.Int64PointerValue(updateCommon.OauthConfigurationId)
		plan.AdapterVersion = types.StringValue(warehouseConfigChanges.AdapterVersion())

	case plan.BigQueryConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.BigQueryConfig](r.client)

		warehouseConfigChanges := dbt_cloud.BigQueryConfig{}

		// BigQuery specific ones
		if plan.BigQueryConfig.GCPProjectID != state.BigQueryConfig.GCPProjectID {
			warehouseConfigChanges.ProjectID = plan.BigQueryConfig.GCPProjectID.ValueStringPointer()
		}
		if plan.BigQueryConfig.TimeoutSeconds != state.BigQueryConfig.TimeoutSeconds {
			warehouseConfigChanges.TimeoutSeconds = plan.BigQueryConfig.TimeoutSeconds.ValueInt64Pointer()
		}
		if plan.BigQueryConfig.PrivateKeyID != state.BigQueryConfig.PrivateKeyID {
			warehouseConfigChanges.PrivateKeyID = plan.BigQueryConfig.PrivateKeyID.ValueStringPointer()
		}
		if plan.BigQueryConfig.PrivateKey != state.BigQueryConfig.PrivateKey {
			warehouseConfigChanges.PrivateKey = plan.BigQueryConfig.PrivateKey.ValueStringPointer()
		}
		if plan.BigQueryConfig.ClientEmail != state.BigQueryConfig.ClientEmail {
			warehouseConfigChanges.ClientEmail = plan.BigQueryConfig.ClientEmail.ValueStringPointer()
		}
		if plan.BigQueryConfig.ClientID != state.BigQueryConfig.ClientID {
			warehouseConfigChanges.ClientID = plan.BigQueryConfig.ClientID.ValueStringPointer()
		}
		if plan.BigQueryConfig.AuthURI != state.BigQueryConfig.AuthURI {
			warehouseConfigChanges.AuthURI = plan.BigQueryConfig.AuthURI.ValueStringPointer()
		}
		if plan.BigQueryConfig.TokenURI != state.BigQueryConfig.TokenURI {
			warehouseConfigChanges.TokenURI = plan.BigQueryConfig.TokenURI.ValueStringPointer()
		}
		if plan.BigQueryConfig.AuthProviderX509CertURL != state.BigQueryConfig.AuthProviderX509CertURL {
			warehouseConfigChanges.AuthProviderX509CertURL = plan.BigQueryConfig.AuthProviderX509CertURL.ValueStringPointer()
		}
		if plan.BigQueryConfig.ClientX509CertURL != state.BigQueryConfig.ClientX509CertURL {
			warehouseConfigChanges.ClientX509CertURL = plan.BigQueryConfig.ClientX509CertURL.ValueStringPointer()
		}
		if plan.BigQueryConfig.Retries != state.BigQueryConfig.Retries {
			warehouseConfigChanges.Retries = plan.BigQueryConfig.Retries.ValueInt64Pointer()
		}
		left, right := lo.Difference(plan.BigQueryConfig.Scopes, state.BigQueryConfig.Scopes)
		if len(left) > 0 || len(right) > 0 {
			warehouseConfigChanges.Scopes = helper.TypesStringSliceToStringSlice(
				plan.BigQueryConfig.Scopes,
			)
		}

		// nullable fields
		// when the values are Null, we still want to send it as null to the PATCH payload, to remove it, otherwise the omitempty doesn't add it to the payload and it doesn't get updated
		if plan.BigQueryConfig.Priority != state.BigQueryConfig.Priority {
			if plan.BigQueryConfig.Priority.IsNull() {
				warehouseConfigChanges.Priority.SetNull()
			} else {
				warehouseConfigChanges.Priority.Set(plan.BigQueryConfig.Priority.ValueString())
			}
		}
		if plan.BigQueryConfig.Location != state.BigQueryConfig.Location {
			if plan.BigQueryConfig.Location.IsNull() {
				warehouseConfigChanges.Location.SetNull()
			} else {
				warehouseConfigChanges.Location.Set(plan.BigQueryConfig.Location.ValueString())
			}
		}
		if plan.BigQueryConfig.MaximumBytesBilled != state.BigQueryConfig.MaximumBytesBilled {
			if plan.BigQueryConfig.MaximumBytesBilled.IsNull() {
				warehouseConfigChanges.MaximumBytesBilled.SetNull()
			} else {
				warehouseConfigChanges.MaximumBytesBilled.Set(plan.BigQueryConfig.MaximumBytesBilled.ValueInt64())
			}
		}
		if plan.BigQueryConfig.ExecutionProject != state.BigQueryConfig.ExecutionProject {
			if plan.BigQueryConfig.ExecutionProject.IsNull() {
				warehouseConfigChanges.ExecutionProject.SetNull()
			} else {
				warehouseConfigChanges.ExecutionProject.Set(plan.BigQueryConfig.ExecutionProject.ValueString())
			}
		}
		if plan.BigQueryConfig.ImpersonateServiceAccount != state.BigQueryConfig.ImpersonateServiceAccount {
			if plan.BigQueryConfig.ImpersonateServiceAccount.IsNull() {
				warehouseConfigChanges.ImpersonateServiceAccount.SetNull()
			} else {
				warehouseConfigChanges.ImpersonateServiceAccount.Set(
					plan.BigQueryConfig.ImpersonateServiceAccount.ValueString(),
				)
			}
		}
		if plan.BigQueryConfig.JobRetryDeadlineSeconds != state.BigQueryConfig.JobRetryDeadlineSeconds {
			if plan.BigQueryConfig.JobRetryDeadlineSeconds.IsNull() {
				warehouseConfigChanges.JobRetryDeadlineSeconds.SetNull()
			} else {
				warehouseConfigChanges.JobRetryDeadlineSeconds.Set(
					plan.BigQueryConfig.JobRetryDeadlineSeconds.ValueInt64(),
				)
			}
		}
		if plan.BigQueryConfig.JobCreationTimeoutSeconds != state.BigQueryConfig.JobCreationTimeoutSeconds {
			if plan.BigQueryConfig.JobCreationTimeoutSeconds.IsNull() {
				warehouseConfigChanges.JobCreationTimeoutSeconds.SetNull()
			} else {
				warehouseConfigChanges.JobCreationTimeoutSeconds.Set(
					plan.BigQueryConfig.JobCreationTimeoutSeconds.ValueInt64(),
				)
			}
		}
		if plan.BigQueryConfig.ApplicationID != state.BigQueryConfig.ApplicationID {
			if plan.BigQueryConfig.ApplicationID.IsNull() {
				warehouseConfigChanges.ApplicationID.SetNull()
			} else {
				warehouseConfigChanges.ApplicationID.Set(plan.BigQueryConfig.ApplicationID.ValueString())
			}
		}
		if plan.BigQueryConfig.ApplicationSecret != state.BigQueryConfig.ApplicationSecret {
			if plan.BigQueryConfig.ApplicationSecret.IsNull() {
				warehouseConfigChanges.ApplicationSecret.SetNull()
			} else {
				warehouseConfigChanges.ApplicationSecret.Set(plan.BigQueryConfig.ApplicationSecret.ValueString())
			}
		}
		if plan.BigQueryConfig.GcsBucket != state.BigQueryConfig.GcsBucket {
			if plan.BigQueryConfig.GcsBucket.IsNull() {
				warehouseConfigChanges.GcsBucket.SetNull()
			} else {
				warehouseConfigChanges.GcsBucket.Set(plan.BigQueryConfig.GcsBucket.ValueString())
			}
		}
		if plan.BigQueryConfig.DataprocRegion != state.BigQueryConfig.DataprocRegion {
			if plan.BigQueryConfig.DataprocRegion.IsNull() {
				warehouseConfigChanges.DataprocRegion.SetNull()
			} else {
				warehouseConfigChanges.DataprocRegion.Set(plan.BigQueryConfig.DataprocRegion.ValueString())
			}
		}
		if plan.BigQueryConfig.DataprocClusterName != state.BigQueryConfig.DataprocClusterName {
			if plan.BigQueryConfig.DataprocClusterName.IsNull() {
				warehouseConfigChanges.DataprocClusterName.SetNull()
			} else {
				warehouseConfigChanges.DataprocClusterName.Set(
					plan.BigQueryConfig.DataprocClusterName.ValueString(),
				)
			}
		}

		updateCommon, _, err := c.Update(
			state.ID.ValueInt64(),
			globalConfigChanges,
			warehouseConfigChanges,
		)
		if err != nil {
			resp.Diagnostics.AddError("Error updating global connection", err.Error())
			return
		}

		// we set the computed values, no need to do it for ID as we use a PlanModifier with UseStateForUnknown()
		plan.IsSshTunnelEnabled = types.BoolPointerValue(updateCommon.IsSshTunnelEnabled)
		plan.OauthConfigurationId = types.Int64PointerValue(updateCommon.OauthConfigurationId)
		plan.AdapterVersion = types.StringValue(warehouseConfigChanges.AdapterVersion())

	}
	// Set the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

}

func (r *globalConnectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	connectionID, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing the connection ID",
			err.Error(),
		)
		return
	}

	globalConnectionResponse, err := r.client.GetGlobalConnectionAdapter(int64(connectionID))
	if err != nil {
		resp.Diagnostics.AddError("Error getting the connection type", err.Error())
		return
	}

	connectionType := strings.Split(globalConnectionResponse.Data.AdapterVersion, "_")[0]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), int64(connectionID))...)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(
			ctx,
			path.Root(connectionType),
			mappingAdapterEmptyConfig[connectionType],
		)...)
}

func (r *globalConnectionResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
