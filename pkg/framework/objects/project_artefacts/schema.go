package project_artefacts

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Schema implements resource.Resource.
func (p *projectArtefactsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "[Deprecated] Resource for mentioning what jobs are the source of truth for the legacy dbt Docs and dbt Source Freshness pages. dbt Explorer doesn't require this config anymore.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the project artefacts resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.Int64Attribute{
				Description: "Project ID",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"docs_job_id": schema.Int64Attribute{
				Description: "Docs Job ID",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"freshness_job_id": schema.Int64Attribute{
				Description: "Freshness Job ID",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
		},
	}
}
