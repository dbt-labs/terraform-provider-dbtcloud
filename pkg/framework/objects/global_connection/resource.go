package global_connection

import (
	"context"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = &globalConnectionResource{}
	_ resource.ResourceWithConfigure        = &globalConnectionResource{}
	_ resource.ResourceWithImportState      = &globalConnectionResource{}
	_ resource.ResourceWithConfigValidators = &globalConnectionResource{}
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
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)
		state.PrivateLinkEndpointId = types.Int64PointerValue(common.PrivateLinkEndpointId)
		state.OauthConfigurationId = types.Int64PointerValue(common.OauthConfigurationId)

		// snowflake settings
		state.SnowflakeConfig.Account = types.StringPointerValue(snowflakeCfg.Account)
		state.SnowflakeConfig.Database = types.StringPointerValue(snowflakeCfg.Database)
		state.SnowflakeConfig.Warehouse = types.StringPointerValue(snowflakeCfg.Warehouse)
		state.SnowflakeConfig.ClientSessionKeepAlive = types.BoolPointerValue(snowflakeCfg.ClientSessionKeepAlive)
		state.SnowflakeConfig.AllowSso = types.BoolPointerValue(snowflakeCfg.AllowSso)

		// nullable optional fields
		// TODO: decide if it is better to read it as string, *string or nullable.Nullable[string] on the dbt_cloud side
		// in this case role can never be empty so this works but we might have cases where null and empty are different
		state.SnowflakeConfig.Role = types.StringPointerValue(snowflakeCfg.Role)

		// We don't set the sensitive fields when we read because those are secret and never returned by the API
		// sensitive fields: OauthClientID, OauthClientSecret

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
		Name:                  plan.Name.ValueStringPointer(),
		IsSshTunnelEnabled:    plan.IsSshTunnelEnabled.ValueBoolPointer(),
		PrivateLinkEndpointId: helper.TypesInt64ToInt64Pointer(plan.PrivateLinkEndpointId),
		OauthConfigurationId:  helper.TypesInt64ToInt64Pointer(plan.OauthConfigurationId),
	}

	switch {
	case plan.SnowflakeConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.SnowflakeConfig](r.client)

		snowflakeCfg := dbt_cloud.SnowflakeConfig{
			Account:                plan.SnowflakeConfig.Account.ValueStringPointer(),
			Database:               plan.SnowflakeConfig.Database.ValueStringPointer(),
			Warehouse:              plan.SnowflakeConfig.Warehouse.ValueStringPointer(),
			ClientSessionKeepAlive: plan.SnowflakeConfig.ClientSessionKeepAlive.ValueBoolPointer(),
			Role:                   plan.SnowflakeConfig.Role.ValueStringPointer(),
			AllowSso:               plan.SnowflakeConfig.AllowSso.ValueBoolPointer(),
			OauthClientID:          plan.SnowflakeConfig.OauthClientID.ValueStringPointer(),
			OauthClientSecret:      plan.SnowflakeConfig.OauthClientSecret.ValueStringPointer(),
		}

		commonResp, _, err := c.Create(commonCfg, snowflakeCfg)

		if err != nil {
			resp.Diagnostics.AddError("Error creating the connection", err.Error())
			return
		}

		// we set the computed values that don't have any default
		plan.ID = types.Int64PointerValue(commonResp.ID)
		plan.IsSshTunnelEnabled = types.BoolPointerValue(commonResp.IsSshTunnelEnabled)

		// TODO(rest)

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
	if plan.PrivateLinkEndpointId != state.PrivateLinkEndpointId {
		globalConfigChanges.PrivateLinkEndpointId = plan.PrivateLinkEndpointId.ValueInt64Pointer()
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
		// here we need to take care of the null case
		// when Role is Null, we still want to send it as null to the PATCH payload, to remove it, otherwise the omitempty doesn't add it to the payload
		if plan.SnowflakeConfig.Role != state.SnowflakeConfig.Role {
			warehouseConfigChanges.Role = plan.SnowflakeConfig.Role.ValueStringPointer()
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

		updateCommon, _, err := c.Update(state.ID.ValueInt64(), globalConfigChanges, warehouseConfigChanges)
		if err != nil {
			resp.Diagnostics.AddError("Error updating global connection", err.Error())
			return
		}

		// we set the computed values, no need to do it for ID as we use a PlanModifier with UseStateForUnknown()
		plan.IsSshTunnelEnabled = types.BoolPointerValue(updateCommon.IsSshTunnelEnabled)

		// Set the updated state
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	}

}

func (r *globalConnectionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// TODO:for the import we need to pass more than just the ID...
	// Or we just pass the ID but we need to get the type of connection first
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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
