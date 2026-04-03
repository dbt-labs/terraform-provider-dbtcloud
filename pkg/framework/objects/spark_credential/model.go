package spark_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SparkCredentialDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	CredentialID types.Int64  `tfsdk:"credential_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	TargetName   types.String `tfsdk:"target_name"`
	NumThreads   types.Int64  `tfsdk:"num_threads"`
	Schema       types.String `tfsdk:"schema"`
}

type SparkCredentialResourceModel struct {
	ID             types.String `tfsdk:"id"`
	CredentialID   types.Int64  `tfsdk:"credential_id"`
	ProjectID      types.Int64  `tfsdk:"project_id"`
	TargetName     types.String `tfsdk:"target_name"`
	Token          types.String `tfsdk:"token"`
	TokenWo        types.String `tfsdk:"token_wo"`
	TokenWoVersion types.Int64  `tfsdk:"token_wo_version"`
	Schema         types.String `tfsdk:"schema"`
}
