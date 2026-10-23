package analytics_credential

import "github.com/hashicorp/terraform-plugin-framework/types"

// AnalyticsCredentialResourceModel is the Terraform state of an analytics credential
type AnalyticsCredentialResourceModel struct {
	ID             types.Int64      `tfsdk:"id"`
	ConnectionID   types.Int64      `tfsdk:"connection_id"`
	AdapterVersion types.String     `tfsdk:"adapter_version"`
	Snowflake      *SnowflakeModel  `tfsdk:"snowflake"`
	Databricks     *DatabricksModel `tfsdk:"databricks"`
	Redshift       *RedshiftModel   `tfsdk:"redshift"`
	BigQuery       *BigQueryModel   `tfsdk:"bigquery"`
}

// BigQueryModel is the bigquery adapter block, for service account and workload identity
// federation credentials; the private key is write-only
type BigQueryModel struct {
	AuthType                       types.String `tfsdk:"auth_type"`
	ExecutionProject               types.String `tfsdk:"execution_project"`
	PrivateKeyID                   types.String `tfsdk:"private_key_id"`
	PrivateKeyWo                   types.String `tfsdk:"private_key_wo"`
	PrivateKeyWoVersion            types.Int64  `tfsdk:"private_key_wo_version"`
	ClientEmail                    types.String `tfsdk:"client_email"`
	ClientID                       types.String `tfsdk:"client_id"`
	AuthURI                        types.String `tfsdk:"auth_uri"`
	TokenURI                       types.String `tfsdk:"token_uri"`
	AuthProviderX509CertURL        types.String `tfsdk:"auth_provider_x509_cert_url"`
	ClientX509CertURL              types.String `tfsdk:"client_x509_cert_url"`
	WorkloadPoolProviderPath       types.String `tfsdk:"workload_pool_provider_path"`
	ServiceAccountImpersonationURL types.String `tfsdk:"service_account_impersonation_url"`
}

// RedshiftModel is the redshift adapter block; the password is write-only
type RedshiftModel struct {
	Username          types.String `tfsdk:"username"`
	PasswordWo        types.String `tfsdk:"password_wo"`
	PasswordWoVersion types.Int64  `tfsdk:"password_wo_version"`
}

// DatabricksModel is the databricks adapter block; the token is write-only
type DatabricksModel struct {
	TokenWo        types.String `tfsdk:"token_wo"`
	TokenWoVersion types.Int64  `tfsdk:"token_wo_version"`
	Catalog        types.String `tfsdk:"catalog"`
}

// SnowflakeModel is the snowflake adapter block. Secrets exist only as write-only
// attributes, so the state never holds them; only their versions are kept.
type SnowflakeModel struct {
	AuthType                      types.String `tfsdk:"auth_type"`
	User                          types.String `tfsdk:"user"`
	Role                          types.String `tfsdk:"role"`
	Warehouse                     types.String `tfsdk:"warehouse"`
	PasswordWo                    types.String `tfsdk:"password_wo"`
	PasswordWoVersion             types.Int64  `tfsdk:"password_wo_version"`
	PrivateKeyWo                  types.String `tfsdk:"private_key_wo"`
	PrivateKeyWoVersion           types.Int64  `tfsdk:"private_key_wo_version"`
	PrivateKeyPassphraseWo        types.String `tfsdk:"private_key_passphrase_wo"`
	PrivateKeyPassphraseWoVersion types.Int64  `tfsdk:"private_key_passphrase_wo_version"`
}

// configuredBlock returns the name of the adapter block that is set, or "" when none is
func (m *AnalyticsCredentialResourceModel) configuredBlock() string {
	for block, present := range m.blocksPresent() {
		if present {
			return block
		}
	}
	return ""
}

// blocksPresent tells, per adapter block, whether it is set
func (m *AnalyticsCredentialResourceModel) blocksPresent() map[string]bool {
	return map[string]bool{
		blockSnowflake:  m.Snowflake != nil,
		blockDatabricks: m.Databricks != nil,
		blockRedshift:   m.Redshift != nil,
		blockBigQuery:   m.BigQuery != nil,
	}
}
