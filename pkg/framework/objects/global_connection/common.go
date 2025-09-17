package global_connection

import (
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readGeneric(
	client *dbt_cloud.Client,
	state *GlobalConnectionResourceModel,
	adapter string,
) (*GlobalConnectionResourceModel, string, error) {

	connectionID := state.ID.ValueInt64()

	switch {
	case state.SnowflakeConfig != nil || strings.HasPrefix(adapter, "snowflake_"):
		// in case we use it for a datasource, we need to set the Config to not be nil
		if state.SnowflakeConfig == nil {
			state.SnowflakeConfig = &SnowflakeConfig{}
		}

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.SnowflakeConfig](client)

		common, snowflakeCfg, err := c.Get(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				return nil, "removeFromState", nil
			}
			return nil, "", err
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(snowflakeCfg.AdapterVersion())
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)

		// nullable common fields
		if !common.PrivateLinkEndpointId.IsNull() {
			state.PrivateLinkEndpointId = types.StringValue(common.PrivateLinkEndpointId.MustGet())
		} else {
			state.PrivateLinkEndpointId = types.StringNull()
		}
		if !common.OauthConfigurationId.IsNull() {
			state.OauthConfigurationId = types.Int64Value(common.OauthConfigurationId.MustGet())
		} else {
			state.OauthConfigurationId = types.Int64Null()
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

	case state.BigQueryConfig != nil || strings.HasPrefix(adapter, "bigquery_"):
		// in case we use it for a datasource, we need to set the Config to not be nil
		if state.BigQueryConfig == nil {
			state.BigQueryConfig = &BigQueryConfig{}
		}

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.BigQueryConfig](client)

		common, bigqueryCfg, adapterVersion, err := c.GetWithAdapterVersion(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				return nil, "removeFromState", nil
			}
			return nil, "", err
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(adapterVersion)
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)

		// nullable common fields
		if !common.PrivateLinkEndpointId.IsNull() {
			state.PrivateLinkEndpointId = types.StringValue(common.PrivateLinkEndpointId.MustGet())
		} else {
			state.PrivateLinkEndpointId = types.StringNull()
		}
		if !common.OauthConfigurationId.IsNull() {
			state.OauthConfigurationId = types.Int64Value(common.OauthConfigurationId.MustGet())
		} else {
			state.OauthConfigurationId = types.Int64Null()
		}

		// BigQuery settings
		state.BigQueryConfig.GCPProjectID = types.StringPointerValue(bigqueryCfg.ProjectID)

		// if the adapter version is not the legacy one, timeout_seconds is not supported
		if adapterVersion == bigqueryCfg.AdapterVersion() {
			state.BigQueryConfig.TimeoutSeconds = types.Int64PointerValue(bigqueryCfg.TimeoutSeconds)
		}

		var jobExecutionTimeoutSeconds int64
		if bigqueryCfg.JobExecutionTimeoutSeconds.IsSpecified() {
			jobExecutionTimeoutSeconds = bigqueryCfg.JobExecutionTimeoutSeconds.MustGet()
			state.BigQueryConfig.JobExecutionTimeoutSeconds = types.Int64PointerValue(&jobExecutionTimeoutSeconds)
		}

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

	case state.DatabricksConfig != nil || strings.HasPrefix(adapter, "databricks_"):
		// in case we use it for a datasource, we need to set the Config to not be nil
		if state.DatabricksConfig == nil {
			state.DatabricksConfig = &DatabricksConfig{}
		}

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.DatabricksConfig](client)

		common, databricksCfg, err := c.Get(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				return nil, "removeFromState", nil
			}
			return nil, "", err
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(databricksCfg.AdapterVersion())
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)

		// nullable common fields
		if !common.PrivateLinkEndpointId.IsNull() {
			state.PrivateLinkEndpointId = types.StringValue(common.PrivateLinkEndpointId.MustGet())
		} else {
			state.PrivateLinkEndpointId = types.StringNull()
		}
		if !common.OauthConfigurationId.IsNull() {
			state.OauthConfigurationId = types.Int64Value(common.OauthConfigurationId.MustGet())
		} else {
			state.OauthConfigurationId = types.Int64Null()
		}

		// Databricks settings
		state.DatabricksConfig.Host = types.StringPointerValue(databricksCfg.Host)
		state.DatabricksConfig.HTTPPath = types.StringPointerValue(databricksCfg.HTTPPath)

		// nullable optional fields
		if !databricksCfg.Catalog.IsNull() {
			state.DatabricksConfig.Catalog = types.StringValue(databricksCfg.Catalog.MustGet())
		} else {
			state.DatabricksConfig.Catalog = types.StringNull()
		}

		// We don't set the sensitive fields when we read because those are secret and never returned by the API
		// sensitive fields: ClientID, ClientSecret

	case state.RedshiftConfig != nil || strings.HasPrefix(adapter, "redshift_"):
		// in case we use it for a datasource, we need to set the Config to not be nil
		if state.RedshiftConfig == nil {
			state.RedshiftConfig = &RedshiftConfig{}
		}

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.RedshiftConfig](client)

		common, redshiftCfg, err := c.Get(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				return nil, "removeFromState", nil
			}
			return nil, "", err
		}

		sshTunnel, err := c.GetEncryptionsForConnection(connectionID)
		if err != nil {
			return nil, "", err
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(redshiftCfg.AdapterVersion())
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)

		// nullable common fields
		if !common.PrivateLinkEndpointId.IsNull() {
			state.PrivateLinkEndpointId = types.StringValue(common.PrivateLinkEndpointId.MustGet())
		} else {
			state.PrivateLinkEndpointId = types.StringNull()
		}
		if !common.OauthConfigurationId.IsNull() {
			state.OauthConfigurationId = types.Int64Value(common.OauthConfigurationId.MustGet())
		} else {
			state.OauthConfigurationId = types.Int64Null()
		}

		// Redshift settings
		state.RedshiftConfig.HostName = types.StringPointerValue(redshiftCfg.HostName)
		state.RedshiftConfig.Port = types.Int64PointerValue(redshiftCfg.Port)

		// nullable optional fields
		if !redshiftCfg.DBName.IsNull() {
			state.RedshiftConfig.DBName = types.StringValue(redshiftCfg.DBName.MustGet())
		} else {
			state.RedshiftConfig.DBName = types.StringNull()
		}

		// SSH tunnel settings
		if len(*sshTunnel) > 0 {

			state.RedshiftConfig.SSHTunnel = &SSHTunnelConfig{
				ID:        types.Int64PointerValue((*sshTunnel)[0].ID),
				HostName:  types.StringValue((*sshTunnel)[0].HostName),
				Port:      types.Int64Value((*sshTunnel)[0].Port),
				Username:  types.StringValue((*sshTunnel)[0].Username),
				PublicKey: types.StringValue((*sshTunnel)[0].PublicKey),
			}
		}

		// We don't set the sensitive fields when we read because those are secret and never returned by the API
		// sensitive fields: N/A for Redshift

	case state.PostgresConfig != nil || strings.HasPrefix(adapter, "postgres_"):
		// in case we use it for a datasource, we need to set the Config to not be nil
		if state.PostgresConfig == nil {
			state.PostgresConfig = &PostgresConfig{}
		}

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.PostgresConfig](client)

		common, postgresCfg, err := c.Get(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				return nil, "removeFromState", nil
			}
			return nil, "", err
		}

		sshTunnel, err := c.GetEncryptionsForConnection(connectionID)
		if err != nil {
			return nil, "", err
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(postgresCfg.AdapterVersion())
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)

		// nullable common fields
		if !common.PrivateLinkEndpointId.IsNull() {
			state.PrivateLinkEndpointId = types.StringValue(common.PrivateLinkEndpointId.MustGet())
		} else {
			state.PrivateLinkEndpointId = types.StringNull()
		}
		if !common.OauthConfigurationId.IsNull() {
			state.OauthConfigurationId = types.Int64Value(common.OauthConfigurationId.MustGet())
		} else {
			state.OauthConfigurationId = types.Int64Null()
		}

		// Postgres settings
		state.PostgresConfig.HostName = types.StringPointerValue(postgresCfg.HostName)
		state.PostgresConfig.Port = types.Int64PointerValue(postgresCfg.Port)

		// nullable optional fields
		if !postgresCfg.DBName.IsNull() {
			state.PostgresConfig.DBName = types.StringValue(postgresCfg.DBName.MustGet())
		} else {
			state.PostgresConfig.DBName = types.StringNull()
		}

		// SSH tunnel settings
		if len(*sshTunnel) > 0 {

			state.PostgresConfig.SSHTunnel = &SSHTunnelConfig{
				ID:        types.Int64PointerValue((*sshTunnel)[0].ID),
				HostName:  types.StringValue((*sshTunnel)[0].HostName),
				Port:      types.Int64Value((*sshTunnel)[0].Port),
				Username:  types.StringValue((*sshTunnel)[0].Username),
				PublicKey: types.StringValue((*sshTunnel)[0].PublicKey),
			}
		}

		// We don't set the sensitive fields when we read because those are secret and never returned by the API
		// sensitive fields: N/A for Postgres

	case state.FabricConfig != nil || strings.HasPrefix(adapter, "fabric_"):
		// in case we use it for a datasource, we need to set the Config to not be nil
		if state.FabricConfig == nil {
			state.FabricConfig = &FabricConfig{}
		}

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.FabricConfig](client)

		common, fabricCfg, err := c.Get(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				return nil, "removeFromState", nil
			}
			return nil, "", err
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(fabricCfg.AdapterVersion())
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)

		// nullable common fields
		if !common.PrivateLinkEndpointId.IsNull() {
			state.PrivateLinkEndpointId = types.StringValue(common.PrivateLinkEndpointId.MustGet())
		} else {
			state.PrivateLinkEndpointId = types.StringNull()
		}
		if !common.OauthConfigurationId.IsNull() {
			state.OauthConfigurationId = types.Int64Value(common.OauthConfigurationId.MustGet())
		} else {
			state.OauthConfigurationId = types.Int64Null()
		}

		// Fabric settings
		state.FabricConfig.Server = types.StringPointerValue(fabricCfg.Server)
		state.FabricConfig.Port = types.Int64PointerValue(fabricCfg.Port)
		state.FabricConfig.Database = types.StringPointerValue(fabricCfg.Database)
		state.FabricConfig.Retries = types.Int64PointerValue(fabricCfg.Retries)
		state.FabricConfig.LoginTimeout = types.Int64PointerValue(fabricCfg.LoginTimeout)
		state.FabricConfig.QueryTimeout = types.Int64PointerValue(fabricCfg.QueryTimeout)

		// We don't set the sensitive fields when we read because those are secret and never returned by the API
		// sensitive fields: N/A for Fabric

	case state.SynapseConfig != nil || strings.HasPrefix(adapter, "synapse_"):
		// in case we use it for a datasource, we need to set the Config to not be nil
		if state.SynapseConfig == nil {
			state.SynapseConfig = &SynapseConfig{}
		}

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.SynapseConfig](client)

		common, synapseCfg, err := c.Get(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				return nil, "removeFromState", nil
			}
			return nil, "", err
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(synapseCfg.AdapterVersion())
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)

		// nullable common fields
		if !common.PrivateLinkEndpointId.IsNull() {
			state.PrivateLinkEndpointId = types.StringValue(common.PrivateLinkEndpointId.MustGet())
		} else {
			state.PrivateLinkEndpointId = types.StringNull()
		}
		if !common.OauthConfigurationId.IsNull() {
			state.OauthConfigurationId = types.Int64Value(common.OauthConfigurationId.MustGet())
		} else {
			state.OauthConfigurationId = types.Int64Null()
		}

		// Synapse settings
		state.SynapseConfig.Host = types.StringPointerValue(synapseCfg.Host)
		state.SynapseConfig.Port = types.Int64PointerValue(synapseCfg.Port)
		state.SynapseConfig.Database = types.StringPointerValue(synapseCfg.Database)
		state.SynapseConfig.Retries = types.Int64PointerValue(synapseCfg.Retries)
		state.SynapseConfig.LoginTimeout = types.Int64PointerValue(synapseCfg.LoginTimeout)
		state.SynapseConfig.QueryTimeout = types.Int64PointerValue(synapseCfg.QueryTimeout)

		// We don't set the sensitive fields when we read because those are secret and never returned by the API
		// sensitive fields: N/A for Synapse

	case state.StarburstConfig != nil || strings.HasPrefix(adapter, "trino_"):
		// in case we use it for a datasource, we need to set the Config to not be nil
		if state.StarburstConfig == nil {
			state.StarburstConfig = &StarburstConfig{}
		}

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.StarburstConfig](client)

		common, starburstCfg, err := c.Get(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				return nil, "removeFromState", nil
			}
			return nil, "", err
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(starburstCfg.AdapterVersion())
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)

		// nullable common fields
		if !common.PrivateLinkEndpointId.IsNull() {
			state.PrivateLinkEndpointId = types.StringValue(common.PrivateLinkEndpointId.MustGet())
		} else {
			state.PrivateLinkEndpointId = types.StringNull()
		}
		if !common.OauthConfigurationId.IsNull() {
			state.OauthConfigurationId = types.Int64Value(common.OauthConfigurationId.MustGet())
		} else {
			state.OauthConfigurationId = types.Int64Null()
		}

		// Starburst settings
		state.StarburstConfig.Method = types.StringPointerValue(starburstCfg.Method)
		state.StarburstConfig.Host = types.StringPointerValue(starburstCfg.Host)
		state.StarburstConfig.Port = types.Int64PointerValue(starburstCfg.Port)

		// We don't set the sensitive fields when we read because those are secret and never returned by the API
		// sensitive fields: N/A for Starburst

	case state.AthenaConfig != nil || strings.HasPrefix(adapter, "athena_"):
		// in case we use it for a datasource, we need to set the Config to not be nil
		if state.AthenaConfig == nil {
			state.AthenaConfig = &AthenaConfig{}
		}

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.AthenaConfig](client)

		common, athenaCfg, err := c.Get(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				return nil, "removeFromState", nil
			}
			return nil, "", err
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(athenaCfg.AdapterVersion())
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)

		// nullable common fields
		if !common.PrivateLinkEndpointId.IsNull() {
			state.PrivateLinkEndpointId = types.StringValue(common.PrivateLinkEndpointId.MustGet())
		} else {
			state.PrivateLinkEndpointId = types.StringNull()
		}
		if !common.OauthConfigurationId.IsNull() {
			state.OauthConfigurationId = types.Int64Value(common.OauthConfigurationId.MustGet())
		} else {
			state.OauthConfigurationId = types.Int64Null()
		}

		// Athena settings
		state.AthenaConfig.RegionName = types.StringPointerValue(athenaCfg.RegionName)
		state.AthenaConfig.Database = types.StringPointerValue(athenaCfg.Database)
		state.AthenaConfig.S3StagingDir = types.StringPointerValue(athenaCfg.S3StagingDir)

		// nullable optional fields
		if !athenaCfg.WorkGroup.IsNull() {
			state.AthenaConfig.WorkGroup = types.StringValue(athenaCfg.WorkGroup.MustGet())
		} else {
			state.AthenaConfig.WorkGroup = types.StringNull()
		}
		if !athenaCfg.SparkWorkGroup.IsNull() {
			state.AthenaConfig.SparkWorkGroup = types.StringValue(
				athenaCfg.SparkWorkGroup.MustGet(),
			)
		} else {
			state.AthenaConfig.SparkWorkGroup = types.StringNull()
		}
		if !athenaCfg.S3DataDir.IsNull() {
			state.AthenaConfig.S3DataDir = types.StringValue(athenaCfg.S3DataDir.MustGet())
		} else {
			state.AthenaConfig.S3DataDir = types.StringNull()
		}
		if !athenaCfg.S3DataNaming.IsNull() {
			state.AthenaConfig.S3DataNaming = types.StringValue(athenaCfg.S3DataNaming.MustGet())
		} else {
			state.AthenaConfig.S3DataNaming = types.StringNull()
		}
		if !athenaCfg.S3TmpTableDir.IsNull() {
			state.AthenaConfig.S3TmpTableDir = types.StringValue(athenaCfg.S3TmpTableDir.MustGet())
		} else {
			state.AthenaConfig.S3TmpTableDir = types.StringNull()
		}
		if !athenaCfg.PollInterval.IsNull() {
			state.AthenaConfig.PollInterval = types.Int64Value(athenaCfg.PollInterval.MustGet())
		} else {
			state.AthenaConfig.PollInterval = types.Int64Null()
		}
		if !athenaCfg.NumRetries.IsNull() {
			state.AthenaConfig.NumRetries = types.Int64Value(athenaCfg.NumRetries.MustGet())
		} else {
			state.AthenaConfig.NumRetries = types.Int64Null()
		}
		if !athenaCfg.NumBoto3Retries.IsNull() {
			state.AthenaConfig.NumBoto3Retries = types.Int64Value(
				athenaCfg.NumBoto3Retries.MustGet(),
			)
		} else {
			state.AthenaConfig.NumBoto3Retries = types.Int64Null()
		}
		if !athenaCfg.NumIcebergRetries.IsNull() {
			state.AthenaConfig.NumIcebergRetries = types.Int64Value(
				athenaCfg.NumIcebergRetries.MustGet(),
			)
		} else {
			state.AthenaConfig.NumIcebergRetries = types.Int64Null()
		}

		// We don't set the sensitive fields when we read because those are secret and never returned by the API
		// sensitive fields: N/A for Athena

	case state.ApacheSparkConfig != nil || strings.HasPrefix(adapter, "apache_spark_"):
		// in case we use it for a datasource, we need to set the Config to not be nil
		if state.ApacheSparkConfig == nil {
			state.ApacheSparkConfig = &ApacheSparkConfig{}
		}

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.ApacheSparkConfig](client)

		common, sparkCfg, err := c.Get(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				return nil, "removeFromState", nil
			}
			return nil, "", err
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(sparkCfg.AdapterVersion())
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)

		// nullable common fields
		if !common.PrivateLinkEndpointId.IsNull() {
			state.PrivateLinkEndpointId = types.StringValue(common.PrivateLinkEndpointId.MustGet())
		} else {
			state.PrivateLinkEndpointId = types.StringNull()
		}
		if !common.OauthConfigurationId.IsNull() {
			state.OauthConfigurationId = types.Int64Value(common.OauthConfigurationId.MustGet())
		} else {
			state.OauthConfigurationId = types.Int64Null()
		}

		// Spark settings
		state.ApacheSparkConfig.Method = types.StringPointerValue(sparkCfg.Method)
		state.ApacheSparkConfig.Host = types.StringPointerValue(sparkCfg.Host)
		state.ApacheSparkConfig.Port = types.Int64PointerValue(sparkCfg.Port)
		state.ApacheSparkConfig.Cluster = types.StringPointerValue(sparkCfg.Cluster)
		state.ApacheSparkConfig.ConnectTimeout = types.Int64PointerValue(sparkCfg.ConnectTimeout)
		state.ApacheSparkConfig.ConnectRetries = types.Int64PointerValue(sparkCfg.ConnectRetries)

		// nullable optional fields
		if !sparkCfg.Organization.IsNull() {
			state.ApacheSparkConfig.Organization = types.StringValue(
				sparkCfg.Organization.MustGet(),
			)
		} else {
			state.ApacheSparkConfig.Organization = types.StringNull()
		}
		if !sparkCfg.User.IsNull() {
			state.ApacheSparkConfig.User = types.StringValue(sparkCfg.User.MustGet())
		} else {
			state.ApacheSparkConfig.User = types.StringNull()
		}
		if !sparkCfg.Auth.IsNull() {
			state.ApacheSparkConfig.Auth = types.StringValue(sparkCfg.Auth.MustGet())
		} else {
			state.ApacheSparkConfig.Auth = types.StringNull()
		}

		// We don't set the sensitive fields when we read because those are secret and never returned by the API
		// sensitive fields: N/A for Spark

	case state.TeradataConfig != nil || strings.HasPrefix(adapter, "teradata_"):
		if state.TeradataConfig == nil {
			state.TeradataConfig = &TeradataConfig{}
		}

		c := dbt_cloud.NewGlobalConnectionClient[dbt_cloud.TeradataConfig](client)
		common, teradataCfg, err := c.Get(connectionID)
		if err != nil {
			if strings.HasPrefix(err.Error(), "resource-not-found") {
				return nil, "removeFromState", nil
			}
			return nil, "", err
		}

		// global settings
		state.ID = types.Int64PointerValue(common.ID)
		state.AdapterVersion = types.StringValue(teradataCfg.AdapterVersion())
		state.Name = types.StringPointerValue(common.Name)
		state.IsSshTunnelEnabled = types.BoolPointerValue(common.IsSshTunnelEnabled)

		// Teradata settings
		state.TeradataConfig.Host = types.StringPointerValue(teradataCfg.Host)
		state.TeradataConfig.Port = types.StringPointerValue(teradataCfg.Port)
		state.TeradataConfig.TMode = types.StringPointerValue(teradataCfg.TMode)
		state.TeradataConfig.RequestTimeout = types.Int64PointerValue(teradataCfg.RequestTimeout)
		state.TeradataConfig.Retries = types.Int64PointerValue(teradataCfg.Retries)

	default:
		panic("Unknown connection type")
	}

	return state, "", nil
}
