package global_connection

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var supportedGlobalConfigTypes = []string{"bigquery", "snowflake"}

type GlobalConnectionResourceModel struct {
	ID                    types.Int64      `tfsdk:"id"`
	AdapterVersion        types.String     `tfsdk:"adapter_version"`
	Name                  types.String     `tfsdk:"name"`
	IsSshTunnelEnabled    types.Bool       `tfsdk:"is_ssh_tunnel_enabled"`
	PrivateLinkEndpointId types.String     `tfsdk:"private_link_endpoint_id"`
	OauthConfigurationId  types.Int64      `tfsdk:"oauth_configuration_id"`
	SnowflakeConfig       *SnowflakeConfig `tfsdk:"snowflake"`
	BigQueryConfig        *BigQueryConfig  `tfsdk:"bigquery"`
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
	Priority                  types.String `tfsdk:"priority"`
	Location                  types.String `tfsdk:"location"`
	MaximumBytesBilled        types.Int64  `tfsdk:"maximum_bytes_billed"`
	ExecutionProject          types.String `tfsdk:"execution_project"`
	ImpersonateServiceAccount types.String `tfsdk:"impersonate_service_account"`
	JobRetryDeadlineSeconds   types.Int64  `tfsdk:"job_retry_deadline_seconds"`
	JobCreationTimeoutSeconds types.Int64  `tfsdk:"job_creation_timeout_seconds"`
	ApplicationID             types.String `tfsdk:"application_id"`
	ApplicationSecret         types.String `tfsdk:"application_secret"`
	GcsBucket                 types.String `tfsdk:"gcs_bucket"`
	DataprocRegion            types.String `tfsdk:"dataproc_region"`
	DataprocClusterName       types.String `tfsdk:"dataproc_cluster_name"`
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

type DatabricksConfig struct{}

type RedshiftConfig struct{}

type PostgresConfig struct{}

type FabricConfig struct{}

type StarburstConfig struct{}

type SparkConfig struct{}

type SynapseConfig struct{}

type GlobalConnectionDataSourceModel struct {
	// TBD, and do we use the same as the for the Resource model?
}

// func ConvertGlobalConnectionModelToData(
// 	model GlobalConnectionResourceModel,
// ) dbt_cloud.Notification {
// TBD
// }
