package bigquery_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BigqueryCredentialResourceModel is the model for the resource
type BigqueryCredentialResourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	IsActive     types.Bool   `tfsdk:"is_active"`
	Dataset      types.String `tfsdk:"dataset"`
	NumThreads   types.Int64  `tfsdk:"num_threads"`
	ConnectionID types.Int64  `tfsdk:"connection_id"`
}

// BigqueryCredentialDataSourceModel is the model for the data source
type BigqueryCredentialDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	IsActive     types.Bool   `tfsdk:"is_active"`
	Dataset      types.String `tfsdk:"dataset"`
	NumThreads   types.Int64  `tfsdk:"num_threads"`
}
