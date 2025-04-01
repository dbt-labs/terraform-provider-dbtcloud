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

// SnowflakeCredentialResourceModel is the model for the resource
type SnowflakeCredentialResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	CredentialID         types.Int64  `tfsdk:"credential_id"`
	ProjectID            types.Int64  `tfsdk:"project_id"`
	User                 types.String `tfsdk:"user"`
	Password             types.String `tfsdk:"password"`
	AuthType             types.String `tfsdk:"auth_type"`
	Database             types.String `tfsdk:"database"`
	Role                 types.String `tfsdk:"role"`
	Warehouse            types.String `tfsdk:"warehouse"`
	Schema               types.String `tfsdk:"schema"`
	PrivateKey           types.String `tfsdk:"private_key"`
	PrivateKeyPassphrase types.String `tfsdk:"private_key_passphrase"`
	IsActive             types.Bool   `tfsdk:"is_active"`
	NumThreads           types.Int64  `tfsdk:"num_threads"`
}
