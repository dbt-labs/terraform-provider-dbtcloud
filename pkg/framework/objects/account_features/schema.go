package account_features

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *accountFeaturesResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: helper.DocString(
			`Manages dbt Cloud global features at the account level, like Advanced CI. The same feature should not be configured in different resources to avoid conflicts.
		
		When destroying the resource or removing the value for an attribute, the features status will not be changed. Deactivating features will require applying them wih the value set to ~~~false~~~.`,
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the account.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"advanced_ci": schema.BoolAttribute{
				Description: "Whether advanced CI is enabled.",
				Optional:    true,
				Computed:    true,
			},
			"partial_parsing": schema.BoolAttribute{
				Description: "Whether partial parsing is enabled.",
				Optional:    true,
				Computed:    true,
			},
			"repo_caching": schema.BoolAttribute{
				Description: "Whether repository caching is enabled.",
				Optional:    true,
				Computed:    true,
			},
			"ai_features": schema.BoolAttribute{
				Description: "Whether AI features are enabled.",
				Optional:    true,
				Computed:    true,
			},
			"warehouse_cost_visibility": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether warehouse cost visibility is enabled.",
			},
		},
	}
}
