package athena_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AthenaCredentialResourceModel is the model for the resource
type AthenaCredentialResourceModel struct {
	ID                          types.String `tfsdk:"id"`
	CredentialID                types.Int64  `tfsdk:"credential_id"`
	ProjectID                   types.Int64  `tfsdk:"project_id"`
	AWSAccessKeyID              types.String `tfsdk:"aws_access_key_id"`
	AWSAccessKeyIDWo            types.String `tfsdk:"aws_access_key_id_wo"`
	AWSAccessKeyIDWoVersion     types.Int64  `tfsdk:"aws_access_key_id_wo_version"`
	AWSSecretAccessKey          types.String `tfsdk:"aws_secret_access_key"`
	AWSSecretAccessKeyWo        types.String `tfsdk:"aws_secret_access_key_wo"`
	AWSSecretAccessKeyWoVersion types.Int64  `tfsdk:"aws_secret_access_key_wo_version"`
	Schema                      types.String `tfsdk:"schema"`
}

// AthenaCredentialDataSourceModel is the model for the data source
type AthenaCredentialDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	Schema       types.String `tfsdk:"schema"`
}
