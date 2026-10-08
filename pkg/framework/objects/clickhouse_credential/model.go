package clickhouse_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ClickhouseCredentialDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	User         types.String `tfsdk:"user"`
	Schema       types.String `tfsdk:"schema"`
	TargetName   types.String `tfsdk:"target_name"`
	Threads      types.Int64  `tfsdk:"threads"`
}

type ClickhouseCredentialResourceModel struct {
	ID                types.String `tfsdk:"id"`
	CredentialID      types.Int64  `tfsdk:"credential_id"`
	ProjectID         types.Int64  `tfsdk:"project_id"`
	User              types.String `tfsdk:"user"`
	Password          types.String `tfsdk:"password"`
	PasswordWo        types.String `tfsdk:"password_wo"`
	PasswordWoVersion types.Int64  `tfsdk:"password_wo_version"`
	Schema            types.String `tfsdk:"schema"`
	TargetName        types.String `tfsdk:"target_name"`
	Threads           types.Int64  `tfsdk:"threads"`
}
