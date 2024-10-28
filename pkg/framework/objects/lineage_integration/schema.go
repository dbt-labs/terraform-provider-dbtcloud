package lineage_integration

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *lineageIntegrationResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: helper.DocString(`
		Setup lineage integration for dbt Cloud to automatically fetch lineage from external BI tools in dbt Explorer. Currently supports Tableau.

		This resource requires having an environment tagged as production already created for you project.
		`),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Combination of `project_id` and `lineage_integration_id`",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"lineage_integration_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the lineage integration",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.Int64Attribute{
				Required:    true,
				Description: "The dbt Cloud project ID for the integration",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The integration type. Today only 'tableau' is supported",
				Default:     stringdefault.StaticString("tableau"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host": schema.StringAttribute{
				Required:    true,
				Description: "The URL of the BI server (see docs for more details)",
			},
			"site_id": schema.StringAttribute{
				Required:    true,
				Description: "The sitename for the collections of dashboards (see docs for more details)",
			},
			"token_name": schema.StringAttribute{
				Required:    true,
				Description: "The token to use to authenticate to the BI server",
			},
			"token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The secret token value to use to authenticate to the BI server",
			},
		},
	}
}
