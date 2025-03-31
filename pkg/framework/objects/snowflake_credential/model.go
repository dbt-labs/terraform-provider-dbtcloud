package snowflake_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SnowflakeCredentialDataSourceModel is the model for the data source
type SnowflakeCredentialDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	IsActive     types.Bool   `tfsdk:"is_active"`
	AuthType     types.String `tfsdk:"auth_type"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	Schema       types.String `tfsdk:"schema"`
	User         types.String `tfsdk:"user"`
	NumThreads   types.Int64  `tfsdk:"num_threads"`
}
