package athena_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AthenaCredentialResourceModel is the model for the resource
type AthenaCredentialResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	CredentialID       types.Int64  `tfsdk:"credential_id"`
	ProjectID          types.Int64  `tfsdk:"project_id"`
	AWSAccessKeyID     types.String `tfsdk:"aws_access_key_id"`
	AWSSecretAccessKey types.String `tfsdk:"aws_secret_access_key"`
	Schema             types.String `tfsdk:"schema"`
	AdapterVersion     types.String `tfsdk:"adapter_version"`
}

// AthenaCredentialDataSourceModel is the model for the data source
type AthenaCredentialDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	Schema       types.String `tfsdk:"schema"`
}
