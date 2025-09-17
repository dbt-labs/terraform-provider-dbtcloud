package global_connection

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

type ConfigDetails struct {
	EmptyConfigName    interface{}
	IsEmptyConfig      func(*GlobalConnectionResourceModel) bool
	GetSSHTunnelConfig func(*GlobalConnectionResourceModel) *SSHTunnelConfig
}

var mappingAdapterDetails = map[string]ConfigDetails{
	"bigquery": {
		EmptyConfigName: BigQueryConfig{},
		IsEmptyConfig: func(model *GlobalConnectionResourceModel) bool {
			return model.BigQueryConfig == nil
		},
		GetSSHTunnelConfig: func(model *GlobalConnectionResourceModel) *SSHTunnelConfig {
			return nil
		},
	},
	"snowflake": {
		EmptyConfigName: SnowflakeConfig{},
		IsEmptyConfig: func(model *GlobalConnectionResourceModel) bool {
			return model.SnowflakeConfig == nil
		},
		GetSSHTunnelConfig: func(model *GlobalConnectionResourceModel) *SSHTunnelConfig {
			return nil
		},
	},
	"databricks": {
		EmptyConfigName: DatabricksConfig{},
		IsEmptyConfig: func(model *GlobalConnectionResourceModel) bool {
			return model.DatabricksConfig == nil
		},
		GetSSHTunnelConfig: func(model *GlobalConnectionResourceModel) *SSHTunnelConfig {
			return nil
		},
	},
	"redshift": {
		EmptyConfigName: RedshiftConfig{},
		IsEmptyConfig: func(model *GlobalConnectionResourceModel) bool {
			return model.RedshiftConfig == nil
		},
		GetSSHTunnelConfig: func(model *GlobalConnectionResourceModel) *SSHTunnelConfig {
			if model.RedshiftConfig != nil {
				return model.RedshiftConfig.SSHTunnel
			} else {
				return nil
			}
		},
	},
	"postgres": {
		EmptyConfigName: PostgresConfig{},
		IsEmptyConfig: func(model *GlobalConnectionResourceModel) bool {
			return model.PostgresConfig == nil
		},
		GetSSHTunnelConfig: func(model *GlobalConnectionResourceModel) *SSHTunnelConfig {
			if model.PostgresConfig != nil {
				return model.PostgresConfig.SSHTunnel
			} else {
				return nil
			}
		},
	},
	"fabric": {
		EmptyConfigName: FabricConfig{},
		IsEmptyConfig: func(model *GlobalConnectionResourceModel) bool {
			return model.FabricConfig == nil
		},
		GetSSHTunnelConfig: func(model *GlobalConnectionResourceModel) *SSHTunnelConfig {
			return nil
		},
	},
	"synapse": {
		EmptyConfigName: SynapseConfig{},
		IsEmptyConfig: func(model *GlobalConnectionResourceModel) bool {
			return model.SynapseConfig == nil
		},
		GetSSHTunnelConfig: func(model *GlobalConnectionResourceModel) *SSHTunnelConfig {
			return nil
		},
	},
	"starburst": {
		EmptyConfigName: StarburstConfig{},
		IsEmptyConfig: func(model *GlobalConnectionResourceModel) bool {
			return model.StarburstConfig == nil
		},
		GetSSHTunnelConfig: func(model *GlobalConnectionResourceModel) *SSHTunnelConfig {
			return nil
		},
	},
	"athena": {
		EmptyConfigName: AthenaConfig{},
		IsEmptyConfig: func(model *GlobalConnectionResourceModel) bool {
			return model.AthenaConfig == nil
		},
		GetSSHTunnelConfig: func(model *GlobalConnectionResourceModel) *SSHTunnelConfig {
			return nil
		},
	},
	"apache_spark": {
		EmptyConfigName: ApacheSparkConfig{},
		IsEmptyConfig: func(model *GlobalConnectionResourceModel) bool {
			return model.ApacheSparkConfig == nil
		},
		GetSSHTunnelConfig: func(model *GlobalConnectionResourceModel) *SSHTunnelConfig {
			return nil
		},
	},
	"teradata": {
		EmptyConfigName: TeradataConfig{},
		IsEmptyConfig: func(model *GlobalConnectionResourceModel) bool {
			return model.TeradataConfig == nil
		},
		GetSSHTunnelConfig: func(model *GlobalConnectionResourceModel) *SSHTunnelConfig {
			return nil
		},
	},
}

var supportedGlobalConfigTypes = lo.Keys(mappingAdapterDetails)

type GlobalConnectionResourceModel struct {
	ID                    types.Int64        `tfsdk:"id"`
	AdapterVersion        types.String       `tfsdk:"adapter_version"`
	Name                  types.String       `tfsdk:"name"`
	IsSshTunnelEnabled    types.Bool         `tfsdk:"is_ssh_tunnel_enabled"` //TODO: check if we can deprecate this
	PrivateLinkEndpointId types.String       `tfsdk:"private_link_endpoint_id"`
	OauthConfigurationId  types.Int64        `tfsdk:"oauth_configuration_id"`
	SnowflakeConfig       *SnowflakeConfig   `tfsdk:"snowflake"`
	BigQueryConfig        *BigQueryConfig    `tfsdk:"bigquery"`
	DatabricksConfig      *DatabricksConfig  `tfsdk:"databricks"`
	RedshiftConfig        *RedshiftConfig    `tfsdk:"redshift"`
	PostgresConfig        *PostgresConfig    `tfsdk:"postgres"`
	FabricConfig          *FabricConfig      `tfsdk:"fabric"`
	SynapseConfig         *SynapseConfig     `tfsdk:"synapse"`
	StarburstConfig       *StarburstConfig   `tfsdk:"starburst"`
	AthenaConfig          *AthenaConfig      `tfsdk:"athena"`
	ApacheSparkConfig     *ApacheSparkConfig `tfsdk:"apache_spark"`
	TeradataConfig        *TeradataConfig    `tfsdk:"teradata"`
}

type SSHTunnelConfig struct {
	ID        types.Int64  `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Port      types.Int64  `tfsdk:"port"`
	HostName  types.String `tfsdk:"hostname"`
	PublicKey types.String `tfsdk:"public_key"`
}

type BigQueryConfig struct {
	GCPProjectID            types.String   `tfsdk:"gcp_project_id"`
	TimeoutSeconds          types.Int64    `tfsdk:"timeout_seconds"`
	PrivateKeyID            types.String   `tfsdk:"private_key_id"`
	PrivateKey              types.String   `tfsdk:"private_key"`
	ClientEmail             types.String   `tfsdk:"client_email"`
	ClientID                types.String   `tfsdk:"client_id"`
	AuthURI                 types.String   `tfsdk:"auth_uri"`
	TokenURI                types.String   `tfsdk:"token_uri"`
	AuthProviderX509CertURL types.String   `tfsdk:"auth_provider_x509_cert_url"`
	ClientX509CertURL       types.String   `tfsdk:"client_x509_cert_url"`
	Retries                 types.Int64    `tfsdk:"retries"`
	Scopes                  []types.String `tfsdk:"scopes"`
	// nullable
	Priority                   types.String `tfsdk:"priority"`
	Location                   types.String `tfsdk:"location"`
	MaximumBytesBilled         types.Int64  `tfsdk:"maximum_bytes_billed"`
	ExecutionProject           types.String `tfsdk:"execution_project"`
	ImpersonateServiceAccount  types.String `tfsdk:"impersonate_service_account"`
	JobRetryDeadlineSeconds    types.Int64  `tfsdk:"job_retry_deadline_seconds"`
	JobCreationTimeoutSeconds  types.Int64  `tfsdk:"job_creation_timeout_seconds"`
	ApplicationID              types.String `tfsdk:"application_id"`
	ApplicationSecret          types.String `tfsdk:"application_secret"`
	GcsBucket                  types.String `tfsdk:"gcs_bucket"`
	DataprocRegion             types.String `tfsdk:"dataproc_region"`
	DataprocClusterName        types.String `tfsdk:"dataproc_cluster_name"`
	UseLegacyAdapter           types.Bool   `tfsdk:"use_legacy_adapter"`
	JobExecutionTimeoutSeconds types.Int64  `tfsdk:"job_execution_timeout_seconds"`
}

type SnowflakeConfig struct {
	Account                types.String `tfsdk:"account"`
	Database               types.String `tfsdk:"database"`
	Warehouse              types.String `tfsdk:"warehouse"`
	ClientSessionKeepAlive types.Bool   `tfsdk:"client_session_keep_alive"`
	AllowSso               types.Bool   `tfsdk:"allow_sso"`
	OauthClientID          types.String `tfsdk:"oauth_client_id"`
	OauthClientSecret      types.String `tfsdk:"oauth_client_secret"`
	// nullable
	Role types.String `tfsdk:"role"`
}

type DatabricksConfig struct {
	Host     types.String `tfsdk:"host"`
	HTTPPath types.String `tfsdk:"http_path"`
	// nullable
	Catalog      types.String `tfsdk:"catalog"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

type RedshiftConfig struct {
	HostName types.String `tfsdk:"hostname"`
	Port     types.Int64  `tfsdk:"port"`
	// nullable
	DBName    types.String     `tfsdk:"dbname"`
	SSHTunnel *SSHTunnelConfig `tfsdk:"ssh_tunnel"`
}

type PostgresConfig struct {
	HostName types.String `tfsdk:"hostname"`
	Port     types.Int64  `tfsdk:"port"`
	// nullable
	DBName    types.String     `tfsdk:"dbname"`
	SSHTunnel *SSHTunnelConfig `tfsdk:"ssh_tunnel"`
}

type FabricConfig struct {
	Server       types.String `tfsdk:"server"`
	Port         types.Int64  `tfsdk:"port"`
	Database     types.String `tfsdk:"database"`
	Retries      types.Int64  `tfsdk:"retries"`
	LoginTimeout types.Int64  `tfsdk:"login_timeout"`
	QueryTimeout types.Int64  `tfsdk:"query_timeout"`
}

// Fabric and Synapse are very similar, except Synapse uses Host instead of Server
type SynapseConfig struct {
	Host         types.String `tfsdk:"host"`
	Port         types.Int64  `tfsdk:"port"`
	Database     types.String `tfsdk:"database"`
	Retries      types.Int64  `tfsdk:"retries"`
	LoginTimeout types.Int64  `tfsdk:"login_timeout"`
	QueryTimeout types.Int64  `tfsdk:"query_timeout"`
}

type StarburstConfig struct {
	Method types.String `tfsdk:"method"`
	Host   types.String `tfsdk:"host"`
	Port   types.Int64  `tfsdk:"port"`
}

type AthenaConfig struct {
	RegionName   types.String `tfsdk:"region_name"`
	Database     types.String `tfsdk:"database"`
	S3StagingDir types.String `tfsdk:"s3_staging_dir"`
	// nullable
	WorkGroup         types.String `tfsdk:"work_group"`
	SparkWorkGroup    types.String `tfsdk:"spark_work_group"`
	S3DataDir         types.String `tfsdk:"s3_data_dir"`
	S3DataNaming      types.String `tfsdk:"s3_data_naming"`
	S3TmpTableDir     types.String `tfsdk:"s3_tmp_table_dir"`
	PollInterval      types.Int64  `tfsdk:"poll_interval"`
	NumRetries        types.Int64  `tfsdk:"num_retries"`
	NumBoto3Retries   types.Int64  `tfsdk:"num_boto3_retries"`
	NumIcebergRetries types.Int64  `tfsdk:"num_iceberg_retries"`
}

type ApacheSparkConfig struct {
	Method         types.String `tfsdk:"method"`
	Host           types.String `tfsdk:"host"`
	Port           types.Int64  `tfsdk:"port"`
	Cluster        types.String `tfsdk:"cluster"`
	ConnectTimeout types.Int64  `tfsdk:"connect_timeout"`
	ConnectRetries types.Int64  `tfsdk:"connect_retries"`
	// nullable
	Organization types.String `tfsdk:"organization"`
	User         types.String `tfsdk:"user"`
	Auth         types.String `tfsdk:"auth"`
}

type TeradataConfig struct {
	Port           types.String `tfsdk:"port"`
	TMode          types.String `tfsdk:"tmode"`
	Host           types.String `tfsdk:"host"`
	Retries        types.Int64  `tfsdk:"retries"`
	RequestTimeout types.Int64  `tfsdk:"request_timeout"`
}

type GlobalConnectionsDatasourceModel struct {
	Connections []GlobalConnectionSummary `tfsdk:"connections"`
}

type GlobalConnectionSummary struct {
	ID                    types.Int64  `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
	AdapterVersion        types.String `tfsdk:"adapter_version"`
	PrivateLinkEndpointID types.String `tfsdk:"private_link_endpoint_id"`
	IsSSHTunnelEnabled    types.Bool   `tfsdk:"is_ssh_tunnel_enabled"`
	OauthConfigurationID  types.Int64  `tfsdk:"oauth_configuration_id"`
	EnvironmentCount      types.Int64  `tfsdk:"environment__count"`
}
