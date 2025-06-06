package databricks_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatabricksCredentialDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	AdapterType  types.String `tfsdk:"adapter_type"`
	TargetName   types.String `tfsdk:"target_name"`
	NumThreads   types.Int64  `tfsdk:"num_threads"`
	Catalog      types.String `tfsdk:"catalog"`
	Schema       types.String `tfsdk:"schema"`
}

type DatabricksCredentialResourceModel struct {
	ID             types.String `tfsdk:"id"`
	CredentialID   types.Int64  `tfsdk:"credential_id"`
	ProjectID      types.Int64  `tfsdk:"project_id"`
	TargetName     types.String `tfsdk:"target_name"`
	Token          types.String `tfsdk:"token"`
	Catalog        types.String `tfsdk:"catalog"`
	Schema         types.String `tfsdk:"schema"`
	AdapterType    types.String `tfsdk:"adapter_type"`
	AdapterVersion types.String `tfsdk:"adapter_version"`
}
