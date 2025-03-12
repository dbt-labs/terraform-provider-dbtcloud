package starburst_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StarburstCredentialResourceModel is the model for the resource
type StarburstCredentialResourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	User         types.String `tfsdk:"user"`
	Password     types.String `tfsdk:"password"`
	Database     types.String `tfsdk:"database"`
	Schema       types.String `tfsdk:"schema"`
}

// StarburstCredentialDataSourceModel is the model for the data source
type StarburstCredentialDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	Database     types.String `tfsdk:"database"`
	Schema       types.String `tfsdk:"schema"`
}
