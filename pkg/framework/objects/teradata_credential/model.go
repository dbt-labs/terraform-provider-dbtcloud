package teradata_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TeradataCredentialModel is the model for the resource
type TeradataCredentialModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	User         types.String `tfsdk:"user"`
	Password     types.String `tfsdk:"password"`
	Schema       types.String `tfsdk:"schema"`
	Threads      types.Int64  `tfsdk:"threads"`
}
