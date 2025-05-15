package semantic_layer_credential

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/bigquery_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/snowflake_credential"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SemanticLayerConfigurationModel struct {
	Name           types.String `tfsdk:"name"`
	AdapterVersion types.String `tfsdk:"adapter_version"`
	ProjectID      types.Int64  `tfsdk:"project_id"`
}

type SnowflakeSLCredentialModel struct {
	ID            types.Int64                                           `tfsdk:"id"`
	Configuration SemanticLayerConfigurationModel                       `tfsdk:"configuration"`
	Credential    snowflake_credential.SnowflakeCredentialResourceModel `tfsdk:"credential"`
}

type BigQuerySLCredentialModel struct {
	ID                  types.Int64                                         `tfsdk:"id"`
	Configuration       SemanticLayerConfigurationModel                     `tfsdk:"configuration"`
	Credential          bigquery_credential.BigqueryCredentialResourceModel `tfsdk:"credential"`
	PrivateKeyID        types.String                                        `tfsdk:"private_key_id"`
	PrivateKey          types.String                                        `tfsdk:"private_key"`
	ClientEmail         types.String                                        `tfsdk:"client_email"`
	ClientID            types.String                                        `tfsdk:"client_id"`
	AuthURI             types.String                                        `tfsdk:"auth_uri"`
	TokenURI            types.String                                        `tfsdk:"token_uri"`
	AuthProviderCertURL types.String                                        `tfsdk:"auth_provider_x509_cert_url"`
	ClientCertURL       types.String                                        `tfsdk:"client_x509_cert_url"`
}
