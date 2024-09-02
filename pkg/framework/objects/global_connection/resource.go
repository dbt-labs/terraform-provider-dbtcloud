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

	var validators []path.Expression
	for _, warehouse := range supportedGlobalConfigTypes {
		validators = append(validators, path.MatchRoot(warehouse))
	}

	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(validators...),
		// BigQuery doesn't support Private Link today
		resourcevalidator.Conflicting(
			path.MatchRoot("bigquery"),
			path.MatchRoot("private_link_endpoint_id"),
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

	for configType, configState := range mappingAdapterDetails {
		wasNull := configState.IsEmptyConfig(&state)
		isNull := configState.IsEmptyConfig(&plan)

		if (wasNull && !isNull) ||
			(!wasNull && isNull) {
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

	newState, action, err := readGeneric(r.client, &state, "")
	if err != nil {
		resp.Diagnostics.AddError("Error reading the connection", err.Error())
		return
	}

	if action == "removeFromState" {
		resp.Diagnostics.AddWarning(
			"Resource not found",
			"The connection resource was not found and has been removed from the state.",
		)
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)

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

	case plan.DatabricksConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.DatabricksConfig](r.client)

		databricksCfg := dbt_cloud.DatabricksConfig{
			Host:     plan.DatabricksConfig.Host.ValueStringPointer(),
			HTTPPath: plan.DatabricksConfig.HTTPPath.ValueStringPointer(),
		}

		// nullable fields
		if !plan.DatabricksConfig.Catalog.IsNull() {
			databricksCfg.Catalog.Set(plan.DatabricksConfig.Catalog.ValueString())
		}
		if !plan.DatabricksConfig.ClientID.IsNull() {
			databricksCfg.ClientID.Set(plan.DatabricksConfig.ClientID.ValueString())
		}
		if !plan.DatabricksConfig.ClientSecret.IsNull() {
			databricksCfg.ClientSecret.Set(plan.DatabricksConfig.ClientSecret.ValueString())
		}

		commonResp, _, err := c.Create(commonCfg, databricksCfg)

		if err != nil {
			resp.Diagnostics.AddError("Error creating the connection", err.Error())
			return
		}

		// we set the computed values that don't have any default
		plan.ID = types.Int64PointerValue(commonResp.ID)
		plan.AdapterVersion = types.StringValue(databricksCfg.AdapterVersion())
		plan.OauthConfigurationId = types.Int64PointerValue(commonResp.OauthConfigurationId)
		plan.IsSshTunnelEnabled = types.BoolPointerValue(commonResp.IsSshTunnelEnabled)

	case plan.RedshiftConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.RedshiftConfig](r.client)

		redshiftCfg := dbt_cloud.RedshiftConfig{
			HostName: plan.RedshiftConfig.HostName.ValueStringPointer(),
			Port:     plan.RedshiftConfig.Port.ValueInt64Pointer(),
		}

		// nullable fields
		if !plan.RedshiftConfig.DBName.IsNull() {
			redshiftCfg.DBName.Set(plan.RedshiftConfig.DBName.ValueString())
		}

		commonResp, _, err := c.Create(commonCfg, redshiftCfg)
		if err != nil {
			resp.Diagnostics.AddError("Error creating the connection", err.Error())
			return
		}

		// SSH tunnel settings
		if plan.RedshiftConfig.SSHTunnel != nil {
			sshTunnelPayload := dbt_cloud.GlobalConnectionEncryptionPayload{
				AccountID:    int64(r.client.AccountID),
				ConnectionID: *commonResp.ID,
				Username:     plan.RedshiftConfig.SSHTunnel.Username.ValueString(),
				Port:         plan.RedshiftConfig.SSHTunnel.Port.ValueInt64(),
				HostName:     plan.RedshiftConfig.SSHTunnel.HostName.ValueString(),
			}
			sshTunnel, err := c.CreateUpdateEncryption(sshTunnelPayload)

			if err != nil {
				resp.Diagnostics.AddError("Error creating the SSH Tunnel", err.Error())
				return
			}

			plan.RedshiftConfig.SSHTunnel = &SSHTunnelConfig{
				ID:        types.Int64PointerValue(sshTunnel.ID),
				Username:  types.StringValue(sshTunnel.Username),
				Port:      types.Int64Value(sshTunnel.Port),
				HostName:  types.StringValue(sshTunnel.HostName),
				PublicKey: types.StringValue(sshTunnel.PublicKey),
			}
		}

		// we set the computed values that don't have any default
		plan.ID = types.Int64PointerValue(commonResp.ID)
		plan.AdapterVersion = types.StringValue(redshiftCfg.AdapterVersion())
		plan.OauthConfigurationId = types.Int64PointerValue(commonResp.OauthConfigurationId)
		plan.IsSshTunnelEnabled = types.BoolPointerValue(commonResp.IsSshTunnelEnabled)

	case plan.PostgresConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.PostgresConfig](r.client)

		postgresCfg := dbt_cloud.PostgresConfig{
			HostName: plan.PostgresConfig.HostName.ValueStringPointer(),
			Port:     plan.PostgresConfig.Port.ValueInt64Pointer(),
		}

		// nullable fields
		if !plan.PostgresConfig.DBName.IsNull() {
			postgresCfg.DBName.Set(plan.PostgresConfig.DBName.ValueString())
		} else {
			// this is different from Redshift. We need to send null on Create for Postgres
			postgresCfg.DBName.SetNull()
		}

		commonResp, _, err := c.Create(commonCfg, postgresCfg)
		if err != nil {
			resp.Diagnostics.AddError("Error creating the connection", err.Error())
			return
		}

		// SSH tunnel settings
		if plan.PostgresConfig.SSHTunnel != nil {
			sshTunnelPayload := dbt_cloud.GlobalConnectionEncryptionPayload{
				AccountID:    int64(r.client.AccountID),
				ConnectionID: *commonResp.ID,
				Username:     plan.PostgresConfig.SSHTunnel.Username.ValueString(),
				Port:         plan.PostgresConfig.SSHTunnel.Port.ValueInt64(),
				HostName:     plan.PostgresConfig.SSHTunnel.HostName.ValueString(),
			}
			sshTunnel, err := c.CreateUpdateEncryption(sshTunnelPayload)

			if err != nil {
				resp.Diagnostics.AddError("Error creating the SSH Tunnel", err.Error())
				return
			}

			plan.PostgresConfig.SSHTunnel = &SSHTunnelConfig{
				ID:        types.Int64PointerValue(sshTunnel.ID),
				Username:  types.StringValue(sshTunnel.Username),
				Port:      types.Int64Value(sshTunnel.Port),
				HostName:  types.StringValue(sshTunnel.HostName),
				PublicKey: types.StringValue(sshTunnel.PublicKey),
			}
		}

		// we set the computed values that don't have any default
		plan.ID = types.Int64PointerValue(commonResp.ID)
		plan.AdapterVersion = types.StringValue(postgresCfg.AdapterVersion())
		plan.OauthConfigurationId = types.Int64PointerValue(commonResp.OauthConfigurationId)
		plan.IsSshTunnelEnabled = types.BoolPointerValue(commonResp.IsSshTunnelEnabled)

	case plan.FabricConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.FabricConfig](r.client)

		fabricCfg := dbt_cloud.FabricConfig{
			Driver:       &dbt_cloud.FabricDriver,
			Server:       plan.FabricConfig.Server.ValueStringPointer(),
			Port:         plan.FabricConfig.Port.ValueInt64Pointer(),
			Database:     plan.FabricConfig.Database.ValueStringPointer(),
			Retries:      plan.FabricConfig.Retries.ValueInt64Pointer(),
			LoginTimeout: plan.FabricConfig.LoginTimeout.ValueInt64Pointer(),
			QueryTimeout: plan.FabricConfig.QueryTimeout.ValueInt64Pointer(),
		}

		// nullable fields
		// N/A for Fabric

		commonResp, _, err := c.Create(commonCfg, fabricCfg)

		if err != nil {
			resp.Diagnostics.AddError("Error creating the connection", err.Error())
			return
		}

		// we set the computed values that don't have any default
		plan.ID = types.Int64PointerValue(commonResp.ID)
		plan.AdapterVersion = types.StringValue(fabricCfg.AdapterVersion())
		plan.OauthConfigurationId = types.Int64PointerValue(commonResp.OauthConfigurationId)
		plan.IsSshTunnelEnabled = types.BoolPointerValue(commonResp.IsSshTunnelEnabled)

	case plan.SynapseConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.SynapseConfig](r.client)

		synapseCfg := dbt_cloud.SynapseConfig{
			Driver:       &dbt_cloud.SynapseDriver,
			Host:         plan.SynapseConfig.Host.ValueStringPointer(),
			Port:         plan.SynapseConfig.Port.ValueInt64Pointer(),
			Database:     plan.SynapseConfig.Database.ValueStringPointer(),
			Retries:      plan.SynapseConfig.Retries.ValueInt64Pointer(),
			LoginTimeout: plan.SynapseConfig.LoginTimeout.ValueInt64Pointer(),
			QueryTimeout: plan.SynapseConfig.QueryTimeout.ValueInt64Pointer(),
		}

		// nullable fields
		// N/A for Synapse

		commonResp, _, err := c.Create(commonCfg, synapseCfg)

		if err != nil {
			resp.Diagnostics.AddError("Error creating the connection", err.Error())
			return
		}

		// we set the computed values that don't have any default
		plan.ID = types.Int64PointerValue(commonResp.ID)
		plan.AdapterVersion = types.StringValue(synapseCfg.AdapterVersion())
		plan.OauthConfigurationId = types.Int64PointerValue(commonResp.OauthConfigurationId)
		plan.IsSshTunnelEnabled = types.BoolPointerValue(commonResp.IsSshTunnelEnabled)

	case plan.StarburstConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.StarburstConfig](r.client)

		starburstCfg := dbt_cloud.StarburstConfig{
			Method: plan.StarburstConfig.Method.ValueStringPointer(),
			Host:   plan.StarburstConfig.Host.ValueStringPointer(),
			Port:   plan.StarburstConfig.Port.ValueInt64Pointer(),
		}

		// nullable fields
		// N/A for Starburst

		commonResp, _, err := c.Create(commonCfg, starburstCfg)

		if err != nil {
			resp.Diagnostics.AddError("Error creating the connection", err.Error())
			return
		}

		// we set the computed values that don't have any default
		plan.ID = types.Int64PointerValue(commonResp.ID)
		plan.AdapterVersion = types.StringValue(starburstCfg.AdapterVersion())
		plan.OauthConfigurationId = types.Int64PointerValue(commonResp.OauthConfigurationId)
		plan.IsSshTunnelEnabled = types.BoolPointerValue(commonResp.IsSshTunnelEnabled)

	case plan.AthenaConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.AthenaConfig](r.client)

		athenaCfg := dbt_cloud.AthenaConfig{
			RegionName:   plan.AthenaConfig.RegionName.ValueStringPointer(),
			Database:     plan.AthenaConfig.Database.ValueStringPointer(),
			S3StagingDir: plan.AthenaConfig.S3StagingDir.ValueStringPointer(),
		}

		// nullable fields
		if !plan.AthenaConfig.WorkGroup.IsNull() {
			athenaCfg.WorkGroup.Set(plan.AthenaConfig.WorkGroup.ValueString())
		}
		if !plan.AthenaConfig.SparkWorkGroup.IsNull() {
			athenaCfg.SparkWorkGroup.Set(plan.AthenaConfig.SparkWorkGroup.ValueString())
		}
		if !plan.AthenaConfig.S3DataDir.IsNull() {
			athenaCfg.S3DataDir.Set(plan.AthenaConfig.S3DataDir.ValueString())
		}
		if !plan.AthenaConfig.S3DataNaming.IsNull() {
			athenaCfg.S3DataNaming.Set(plan.AthenaConfig.S3DataNaming.ValueString())
		}
		if !plan.AthenaConfig.S3TmpTableDir.IsNull() {
			athenaCfg.S3TmpTableDir.Set(plan.AthenaConfig.S3TmpTableDir.ValueString())
		}
		if !plan.AthenaConfig.PollInterval.IsNull() {
			athenaCfg.PollInterval.Set(plan.AthenaConfig.PollInterval.ValueInt64())
		}
		if !plan.AthenaConfig.NumRetries.IsNull() {
			athenaCfg.NumRetries.Set(plan.AthenaConfig.NumRetries.ValueInt64())
		}
		if !plan.AthenaConfig.NumBoto3Retries.IsNull() {
			athenaCfg.NumBoto3Retries.Set(plan.AthenaConfig.NumBoto3Retries.ValueInt64())
		}
		if !plan.AthenaConfig.NumIcebergRetries.IsNull() {
			athenaCfg.NumIcebergRetries.Set(plan.AthenaConfig.NumIcebergRetries.ValueInt64())
		}

		commonResp, _, err := c.Create(commonCfg, athenaCfg)

		if err != nil {
			resp.Diagnostics.AddError("Error creating the connection", err.Error())
			return
		}

		// we set the computed values that don't have any default
		plan.ID = types.Int64PointerValue(commonResp.ID)
		plan.AdapterVersion = types.StringValue(athenaCfg.AdapterVersion())
		plan.OauthConfigurationId = types.Int64PointerValue(commonResp.OauthConfigurationId)
		plan.IsSshTunnelEnabled = types.BoolPointerValue(commonResp.IsSshTunnelEnabled)

	case plan.ApacheSparkConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.ApacheSparkConfig](r.client)

		sparkCfg := dbt_cloud.ApacheSparkConfig{
			Method:         plan.ApacheSparkConfig.Method.ValueStringPointer(),
			Host:           plan.ApacheSparkConfig.Host.ValueStringPointer(),
			Port:           plan.ApacheSparkConfig.Port.ValueInt64Pointer(),
			Cluster:        plan.ApacheSparkConfig.Cluster.ValueStringPointer(),
			ConnectTimeout: plan.ApacheSparkConfig.ConnectTimeout.ValueInt64Pointer(),
			ConnectRetries: plan.ApacheSparkConfig.ConnectRetries.ValueInt64Pointer(),
		}

		// nullable fields
		// Careful, this seems to be handled differently for Spark vs all the other DWs. Here we need to set the fields to NULL
		if !plan.ApacheSparkConfig.Organization.IsNull() {
			sparkCfg.Organization.Set(plan.ApacheSparkConfig.Organization.ValueString())
		} else {
			sparkCfg.Organization.SetNull()
		}
		if !plan.ApacheSparkConfig.User.IsNull() {
			sparkCfg.User.Set(plan.ApacheSparkConfig.User.ValueString())
		} else {
			sparkCfg.User.SetNull()
		}
		if !plan.ApacheSparkConfig.Auth.IsNull() {
			sparkCfg.Auth.Set(plan.ApacheSparkConfig.Auth.ValueString())
		} else {
			sparkCfg.Auth.SetNull()
		}

		commonResp, _, err := c.Create(commonCfg, sparkCfg)

		if err != nil {
			resp.Diagnostics.AddError("Error creating the connection", err.Error())
			return
		}

		// we set the computed values that don't have any default
		plan.ID = types.Int64PointerValue(commonResp.ID)
		plan.AdapterVersion = types.StringValue(sparkCfg.AdapterVersion())
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

	// delete the SSH Tunnel if it exists
	for _, config := range mappingAdapterDetails {
		SSHTunnelConfig := config.GetSSHTunnelConfig(&state)
		if SSHTunnelConfig != nil {

			valueID := SSHTunnelConfig.ID.ValueInt64()
			// to delete the encryption we update it with state=2
			sshTunnelPayload := dbt_cloud.GlobalConnectionEncryptionPayload{
				ID:           &valueID,
				AccountID:    int64(r.client.AccountID),
				ConnectionID: connectionID,
				Username:     SSHTunnelConfig.Username.ValueString(),
				Port:         SSHTunnelConfig.Port.ValueInt64(),
				HostName:     SSHTunnelConfig.HostName.ValueString(),
				State:        dbt_cloud.STATE_DELETED,
			}

			// we use Redshift here but it is the same function for all
			// we could change the function to use a generic client rather than a global connection client
			c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.RedshiftConfig](r.client)
			_, err := c.CreateUpdateEncryption(sshTunnelPayload)
			if err != nil {
				resp.Diagnostics.AddError("Error deleting the SSH Tunnel", err.Error())
				return
			}
		}
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

	case plan.DatabricksConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.DatabricksConfig](r.client)

		warehouseConfigChanges := dbt_cloud.DatabricksConfig{}

		// Databricks specific ones
		if plan.DatabricksConfig.Host != state.DatabricksConfig.Host {
			warehouseConfigChanges.Host = plan.DatabricksConfig.Host.ValueStringPointer()
		}
		if plan.DatabricksConfig.HTTPPath != state.DatabricksConfig.HTTPPath {
			warehouseConfigChanges.HTTPPath = plan.DatabricksConfig.HTTPPath.ValueStringPointer()
		}

		// nullable fields
		// when the values are Null, we still want to send it as null to the PATCH payload, to remove it, otherwise the omitempty doesn't add it to the payload and it doesn't get updated
		if plan.DatabricksConfig.Catalog != state.DatabricksConfig.Catalog {
			if plan.DatabricksConfig.Catalog.IsNull() {
				warehouseConfigChanges.Catalog.SetNull()
			} else {
				warehouseConfigChanges.Catalog.Set(plan.DatabricksConfig.Catalog.ValueString())
			}
		}
		if plan.DatabricksConfig.ClientID != state.DatabricksConfig.ClientID {
			if plan.DatabricksConfig.ClientID.IsNull() {
				warehouseConfigChanges.ClientID.SetNull()
			} else {
				warehouseConfigChanges.ClientID.Set(plan.DatabricksConfig.ClientID.ValueString())
			}
		}
		if plan.DatabricksConfig.ClientSecret != state.DatabricksConfig.ClientSecret {
			if plan.DatabricksConfig.ClientSecret.IsNull() {
				warehouseConfigChanges.ClientSecret.SetNull()
			} else {
				warehouseConfigChanges.ClientSecret.Set(plan.DatabricksConfig.ClientSecret.ValueString())
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

	case plan.RedshiftConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.RedshiftConfig](r.client)

		warehouseConfigChanges := dbt_cloud.RedshiftConfig{}
		warehouseConfigChanged := false

		// Redshift specific ones
		if plan.RedshiftConfig.HostName != state.RedshiftConfig.HostName {
			warehouseConfigChanges.HostName = plan.RedshiftConfig.HostName.ValueStringPointer()
			warehouseConfigChanged = true
		}
		if plan.RedshiftConfig.Port != state.RedshiftConfig.Port {
			warehouseConfigChanges.Port = plan.RedshiftConfig.Port.ValueInt64Pointer()
			warehouseConfigChanged = true
		}

		// nullable fields
		// when the values are Null, we still want to send it as null to the PATCH payload, to remove it, otherwise the omitempty doesn't add it to the payload and it doesn't get updated
		if plan.RedshiftConfig.DBName != state.RedshiftConfig.DBName {
			warehouseConfigChanged = true
			if plan.RedshiftConfig.DBName.IsNull() {
				warehouseConfigChanges.DBName.SetNull()
			} else {
				warehouseConfigChanges.DBName.Set(plan.RedshiftConfig.DBName.ValueString())
			}
		}

		if warehouseConfigChanged {
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
		} else {
			// if the warehouseConfig didn't change, we keep the existing state values
			plan.IsSshTunnelEnabled = state.IsSshTunnelEnabled
			plan.OauthConfigurationId = state.OauthConfigurationId
			plan.AdapterVersion = state.AdapterVersion
		}

		// SSH tunnel settings
		sshTunnel, err := r.handleSSHTunnelUpdates(
			plan.RedshiftConfig.SSHTunnel,
			state.RedshiftConfig.SSHTunnel,
			int64(r.client.AccountID),
			state.ID.ValueInt64(),
		)
		if err != nil {
			resp.Diagnostics.AddError("Error with the SSH Tunnel", err.Error())
			return
		}
		plan.RedshiftConfig.SSHTunnel = sshTunnel

	case plan.PostgresConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.PostgresConfig](r.client)

		warehouseConfigChanges := dbt_cloud.PostgresConfig{}
		warehouseConfigChanged := false

		// Postgres specific ones
		if plan.PostgresConfig.HostName != state.PostgresConfig.HostName {
			warehouseConfigChanges.HostName = plan.PostgresConfig.HostName.ValueStringPointer()
			warehouseConfigChanged = true
		}
		if plan.PostgresConfig.Port != state.PostgresConfig.Port {
			warehouseConfigChanges.Port = plan.PostgresConfig.Port.ValueInt64Pointer()
			warehouseConfigChanged = true
		}

		// nullable fields
		// when the values are Null, we still want to send it as null to the PATCH payload, to remove it, otherwise the omitempty doesn't add it to the payload and it doesn't get updated
		if plan.PostgresConfig.DBName != state.PostgresConfig.DBName {
			warehouseConfigChanged = true
			if plan.PostgresConfig.DBName.IsNull() {
				warehouseConfigChanges.DBName.SetNull()
			} else {
				warehouseConfigChanges.DBName.Set(plan.PostgresConfig.DBName.ValueString())
			}
		}

		if warehouseConfigChanged {
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
		} else {
			// if the warehouseConfig didn't change, we keep the existing state values
			plan.IsSshTunnelEnabled = state.IsSshTunnelEnabled
			plan.OauthConfigurationId = state.OauthConfigurationId
			plan.AdapterVersion = state.AdapterVersion
		}

		// SSH tunnel settings
		sshTunnel, err := r.handleSSHTunnelUpdates(
			plan.PostgresConfig.SSHTunnel,
			state.PostgresConfig.SSHTunnel,
			int64(r.client.AccountID),
			state.ID.ValueInt64(),
		)
		if err != nil {
			resp.Diagnostics.AddError("Error with the SSH Tunnel", err.Error())
			return
		}
		plan.PostgresConfig.SSHTunnel = sshTunnel

	case plan.FabricConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.FabricConfig](r.client)

		warehouseConfigChanges := dbt_cloud.FabricConfig{}

		// Fabric specific ones
		if plan.FabricConfig.Server != state.FabricConfig.Server {
			warehouseConfigChanges.Server = plan.FabricConfig.Server.ValueStringPointer()
		}
		if plan.FabricConfig.Port != state.FabricConfig.Port {
			warehouseConfigChanges.Port = plan.FabricConfig.Port.ValueInt64Pointer()
		}
		if plan.FabricConfig.Database != state.FabricConfig.Database {
			warehouseConfigChanges.Database = plan.FabricConfig.Database.ValueStringPointer()
		}
		if plan.FabricConfig.Retries != state.FabricConfig.Retries {
			warehouseConfigChanges.Retries = plan.FabricConfig.Retries.ValueInt64Pointer()
		}
		if plan.FabricConfig.LoginTimeout != state.FabricConfig.LoginTimeout {
			warehouseConfigChanges.LoginTimeout = plan.FabricConfig.LoginTimeout.ValueInt64Pointer()
		}
		if plan.FabricConfig.QueryTimeout != state.FabricConfig.QueryTimeout {
			warehouseConfigChanges.QueryTimeout = plan.FabricConfig.QueryTimeout.ValueInt64Pointer()
		}

		// nullable fields
		// N/A for Fabric

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

	case plan.SynapseConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.SynapseConfig](r.client)

		warehouseConfigChanges := dbt_cloud.SynapseConfig{}

		// Synapse specific ones
		if plan.SynapseConfig.Host != state.SynapseConfig.Host {
			warehouseConfigChanges.Host = plan.SynapseConfig.Host.ValueStringPointer()
		}
		if plan.SynapseConfig.Port != state.SynapseConfig.Port {
			warehouseConfigChanges.Port = plan.SynapseConfig.Port.ValueInt64Pointer()
		}
		if plan.SynapseConfig.Database != state.SynapseConfig.Database {
			warehouseConfigChanges.Database = plan.SynapseConfig.Database.ValueStringPointer()
		}
		if plan.SynapseConfig.Retries != state.SynapseConfig.Retries {
			warehouseConfigChanges.Retries = plan.SynapseConfig.Retries.ValueInt64Pointer()
		}
		if plan.SynapseConfig.LoginTimeout != state.SynapseConfig.LoginTimeout {
			warehouseConfigChanges.LoginTimeout = plan.SynapseConfig.LoginTimeout.ValueInt64Pointer()
		}
		if plan.SynapseConfig.QueryTimeout != state.SynapseConfig.QueryTimeout {
			warehouseConfigChanges.QueryTimeout = plan.SynapseConfig.QueryTimeout.ValueInt64Pointer()
		}

		// nullable fields
		// N/A for Synapse

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

	case plan.StarburstConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.StarburstConfig](r.client)

		warehouseConfigChanges := dbt_cloud.StarburstConfig{}

		// Synapse specific ones
		if plan.StarburstConfig.Method != state.StarburstConfig.Method {
			warehouseConfigChanges.Method = plan.StarburstConfig.Method.ValueStringPointer()
		}
		if plan.StarburstConfig.Host != state.StarburstConfig.Host {
			warehouseConfigChanges.Host = plan.StarburstConfig.Host.ValueStringPointer()
		}
		if plan.StarburstConfig.Port != state.StarburstConfig.Port {
			warehouseConfigChanges.Port = plan.StarburstConfig.Port.ValueInt64Pointer()
		}

		// nullable fields
		// N/A for Starburst

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

	case plan.AthenaConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.AthenaConfig](r.client)

		warehouseConfigChanges := dbt_cloud.AthenaConfig{}

		// Athena specific ones
		if plan.AthenaConfig.RegionName != state.AthenaConfig.RegionName {
			warehouseConfigChanges.RegionName = plan.AthenaConfig.RegionName.ValueStringPointer()
		}
		if plan.AthenaConfig.Database != state.AthenaConfig.Database {
			warehouseConfigChanges.Database = plan.AthenaConfig.Database.ValueStringPointer()
		}
		if plan.AthenaConfig.S3StagingDir != state.AthenaConfig.S3StagingDir {
			warehouseConfigChanges.S3StagingDir = plan.AthenaConfig.S3StagingDir.ValueStringPointer()
		}

		// nullable fields
		// when the values are Null, we still want to send it as null to the PATCH payload, to remove it, otherwise the omitempty doesn't add it to the payload and it doesn't get updated
		if plan.AthenaConfig.WorkGroup != state.AthenaConfig.WorkGroup {
			if plan.AthenaConfig.WorkGroup.IsNull() {
				warehouseConfigChanges.WorkGroup.SetNull()
			} else {
				warehouseConfigChanges.WorkGroup.Set(plan.AthenaConfig.WorkGroup.ValueString())
			}
		}
		if plan.AthenaConfig.SparkWorkGroup != state.AthenaConfig.SparkWorkGroup {
			if plan.AthenaConfig.SparkWorkGroup.IsNull() {
				warehouseConfigChanges.SparkWorkGroup.SetNull()
			} else {
				warehouseConfigChanges.SparkWorkGroup.Set(plan.AthenaConfig.SparkWorkGroup.ValueString())
			}
		}
		if plan.AthenaConfig.S3DataDir != state.AthenaConfig.S3DataDir {
			if plan.AthenaConfig.S3DataDir.IsNull() {
				warehouseConfigChanges.S3DataDir.SetNull()
			} else {
				warehouseConfigChanges.S3DataDir.Set(plan.AthenaConfig.S3DataDir.ValueString())
			}
		}
		if plan.AthenaConfig.S3DataNaming != state.AthenaConfig.S3DataNaming {
			if plan.AthenaConfig.S3DataNaming.IsNull() {
				warehouseConfigChanges.S3DataNaming.SetNull()
			} else {
				warehouseConfigChanges.S3DataNaming.Set(plan.AthenaConfig.S3DataNaming.ValueString())
			}
		}
		if plan.AthenaConfig.S3TmpTableDir != state.AthenaConfig.S3TmpTableDir {
			if plan.AthenaConfig.S3TmpTableDir.IsNull() {
				warehouseConfigChanges.S3TmpTableDir.SetNull()
			} else {
				warehouseConfigChanges.S3TmpTableDir.Set(plan.AthenaConfig.S3TmpTableDir.ValueString())
			}
		}
		if plan.AthenaConfig.PollInterval != state.AthenaConfig.PollInterval {
			if plan.AthenaConfig.PollInterval.IsNull() {
				warehouseConfigChanges.PollInterval.SetNull()
			} else {
				warehouseConfigChanges.PollInterval.Set(plan.AthenaConfig.PollInterval.ValueInt64())
			}
		}
		if plan.AthenaConfig.NumRetries != state.AthenaConfig.NumRetries {
			if plan.AthenaConfig.NumRetries.IsNull() {
				warehouseConfigChanges.NumRetries.SetNull()
			} else {
				warehouseConfigChanges.NumRetries.Set(plan.AthenaConfig.NumRetries.ValueInt64())
			}
		}
		if plan.AthenaConfig.NumBoto3Retries != state.AthenaConfig.NumBoto3Retries {
			if plan.AthenaConfig.NumBoto3Retries.IsNull() {
				warehouseConfigChanges.NumBoto3Retries.SetNull()
			} else {
				warehouseConfigChanges.NumBoto3Retries.Set(plan.AthenaConfig.NumBoto3Retries.ValueInt64())
			}
		}
		if plan.AthenaConfig.NumIcebergRetries != state.AthenaConfig.NumIcebergRetries {
			if plan.AthenaConfig.NumIcebergRetries.IsNull() {
				warehouseConfigChanges.NumIcebergRetries.SetNull()
			} else {
				warehouseConfigChanges.NumIcebergRetries.Set(plan.AthenaConfig.NumIcebergRetries.ValueInt64())
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

	case plan.ApacheSparkConfig != nil:

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.ApacheSparkConfig](r.client)

		warehouseConfigChanges := dbt_cloud.ApacheSparkConfig{}

		// Spark specific ones
		if plan.ApacheSparkConfig.Method != state.ApacheSparkConfig.Method {
			warehouseConfigChanges.Method = plan.ApacheSparkConfig.Method.ValueStringPointer()
		}
		if plan.ApacheSparkConfig.Host != state.ApacheSparkConfig.Host {
			warehouseConfigChanges.Host = plan.ApacheSparkConfig.Host.ValueStringPointer()
		}
		if plan.ApacheSparkConfig.Port != state.ApacheSparkConfig.Port {
			warehouseConfigChanges.Port = plan.ApacheSparkConfig.Port.ValueInt64Pointer()
		}
		if plan.ApacheSparkConfig.Cluster != state.ApacheSparkConfig.Cluster {
			warehouseConfigChanges.Cluster = plan.ApacheSparkConfig.Cluster.ValueStringPointer()
		}
		if plan.ApacheSparkConfig.ConnectTimeout != state.ApacheSparkConfig.ConnectTimeout {
			warehouseConfigChanges.ConnectTimeout = plan.ApacheSparkConfig.ConnectTimeout.ValueInt64Pointer()
		}
		if plan.ApacheSparkConfig.ConnectRetries != state.ApacheSparkConfig.ConnectRetries {
			warehouseConfigChanges.ConnectRetries = plan.ApacheSparkConfig.ConnectRetries.ValueInt64Pointer()
		}

		// nullable fields
		// when the values are Null, we still want to send it as null to the PATCH payload, to remove it, otherwise the omitempty doesn't add it to the payload and it doesn't get updated
		if plan.ApacheSparkConfig.Organization != state.ApacheSparkConfig.Organization {
			if plan.ApacheSparkConfig.Organization.IsNull() {
				warehouseConfigChanges.Organization.SetNull()
			} else {
				warehouseConfigChanges.Organization.Set(plan.ApacheSparkConfig.Organization.ValueString())
			}
		}
		if plan.ApacheSparkConfig.User != state.ApacheSparkConfig.User {
			if plan.ApacheSparkConfig.User.IsNull() {
				warehouseConfigChanges.User.SetNull()
			} else {
				warehouseConfigChanges.User.Set(plan.ApacheSparkConfig.User.ValueString())
			}
		}
		if plan.ApacheSparkConfig.Auth != state.ApacheSparkConfig.Auth {
			if plan.ApacheSparkConfig.Auth.IsNull() {
				warehouseConfigChanges.Auth.SetNull()
			} else {
				warehouseConfigChanges.Auth.Set(plan.ApacheSparkConfig.Auth.ValueString())
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

	default:
		panic("Unknown connection type")
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

	// we need this logic because sometimes adapter names have _ in them, like apache_spark_v0
	var connectionType string
	lastUnderscoreIndex := strings.LastIndex(globalConnectionResponse.Data.AdapterVersion, "_")
	if lastUnderscoreIndex == -1 {
		connectionType = globalConnectionResponse.Data.AdapterVersion
	} else {
		connectionType = globalConnectionResponse.Data.AdapterVersion[:lastUnderscoreIndex]
	}

	// this is the exception where we store the connection details under starburst instead of Trino
	if connectionType == "trino" {
		connectionType = "starburst"
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), int64(connectionID))...)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(
			ctx,
			path.Root(connectionType),
			mappingAdapterDetails[connectionType].EmptyConfigName,
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

func (r *globalConnectionResource) handleSSHTunnelUpdates(
	sshTunnelPlan *SSHTunnelConfig,
	sshTunnelState *SSHTunnelConfig,
	accountID int64,
	connectionID int64,
) (*SSHTunnelConfig, error) {
	c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.RedshiftConfig](r.client)
	if sshTunnelPlan == nil && sshTunnelState != nil {
		// delete the encryption
		valueID := sshTunnelState.ID.ValueInt64()
		sshTunnelPayload := dbt_cloud.GlobalConnectionEncryptionPayload{
			ID:           &valueID,
			AccountID:    accountID,
			ConnectionID: connectionID,
			Username:     sshTunnelState.Username.ValueString(),
			Port:         sshTunnelState.Port.ValueInt64(),
			HostName:     sshTunnelState.HostName.ValueString(),
			State:        dbt_cloud.STATE_DELETED,
		}
		_, err := c.CreateUpdateEncryption(sshTunnelPayload)
		if err != nil {
			return nil, err
		}
	} else if sshTunnelPlan != nil && sshTunnelState == nil {
		// create the encryption
		sshPayload := dbt_cloud.GlobalConnectionEncryptionPayload{
			AccountID:    accountID,
			ConnectionID: connectionID,
			Username:     sshTunnelPlan.Username.ValueString(),
			Port:         sshTunnelPlan.Port.ValueInt64(),
			HostName:     sshTunnelPlan.HostName.ValueString(),
		}
		sshTunnel, err := c.CreateUpdateEncryption(sshPayload)
		if err != nil {
			return nil, err
		}
		sshTunnelPlan = &SSHTunnelConfig{
			ID:        types.Int64PointerValue(sshTunnel.ID),
			Username:  types.StringValue(sshTunnel.Username),
			Port:      types.Int64Value(sshTunnel.Port),
			HostName:  types.StringValue(sshTunnel.HostName),
			PublicKey: types.StringValue(sshTunnel.PublicKey),
		}
	} else if sshTunnelPlan != nil && sshTunnelState != nil && sshTunnelPlan != sshTunnelState {
		// update the encryption
		valueID := sshTunnelState.ID.ValueInt64()
		sshPayload := dbt_cloud.GlobalConnectionEncryptionPayload{
			ID:           &valueID,
			AccountID:    accountID,
			ConnectionID: connectionID,
			Username:     sshTunnelPlan.Username.ValueString(),
			Port:         sshTunnelPlan.Port.ValueInt64(),
			HostName:     sshTunnelPlan.HostName.ValueString(),
		}
		sshTunnel, err := c.CreateUpdateEncryption(sshPayload)
		if err != nil {
			return nil, err
		}
		sshTunnelPlan = &SSHTunnelConfig{
			ID:        types.Int64PointerValue(sshTunnel.ID),
			Username:  types.StringValue(sshTunnel.Username),
			Port:      types.Int64Value(sshTunnel.Port),
			HostName:  types.StringValue(sshTunnel.HostName),
			PublicKey: types.StringValue(sshTunnel.PublicKey),
		}
	}
	return sshTunnelPlan, nil
}
