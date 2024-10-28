package lineage_integration

import "github.com/hashicorp/terraform-plugin-framework/types"

type LineageIntegrationResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	LineageIntegrationID types.Int64  `tfsdk:"lineage_integration_id"`
	ProjectID            types.Int64  `tfsdk:"project_id"`
	Name                 types.String `tfsdk:"name"`
	Host                 types.String `tfsdk:"host"`
	SiteID               types.String `tfsdk:"site_id"`
	TokenName            types.String `tfsdk:"token_name"`
	Token                types.String `tfsdk:"token"`
}
