package azure_dev_ops_repository

import "github.com/hashicorp/terraform-plugin-framework/types"

type AzureDevopsRepositoryDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	AzureDevOpsProjectID types.String `tfsdk:"azure_dev_ops_project_id"`
	DetailsURL           types.String `tfsdk:"details_url"`
	RemoteURL            types.String `tfsdk:"remote_url"`
	WebURL               types.String `tfsdk:"web_url"`
	DefaultBranch        types.String `tfsdk:"default_branch"`
}
