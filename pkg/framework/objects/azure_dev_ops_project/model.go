package azure_dev_ops_project

import "github.com/hashicorp/terraform-plugin-framework/types"

type AzureDevOpsProjectDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	URL  types.String `tfsdk:"url"`
}
