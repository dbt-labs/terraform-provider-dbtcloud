package databricks_credentials

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatabricksCredentialDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	AdapterID    types.Int64  `tfsdk:"adapter_id"`
	TargetName   types.String `tfsdk:"target_name"`
	NumThreads   types.Int64  `tfsdk:"num_threads"`
	Catalog      types.String `tfsdk:"catalog"`
	Schema       types.String `tfsdk:"schema"`
}

type DatabricksCredentialResourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	AdapterID    types.Int64  `tfsdk:"adapter_id"`
	TargetName   types.String `tfsdk:"target_name"`
	Token        types.String `tfsdk:"token"`
	Catalog      types.String `tfsdk:"catalog"`
	Schema       types.String `tfsdk:"schema"`
	AdapterType  types.String `tfsdk:"adapter_type"`
}
