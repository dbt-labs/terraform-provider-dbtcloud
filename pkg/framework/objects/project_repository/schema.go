package project_repository

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// Schema defines the schema for the resource.
func Schema() schema.Schema {
	return schema.Schema{
		Description: "Manages a dbt Cloud project repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the project repository.",
				Computed:    true,
			},
			"repository_id": schema.Int64Attribute{
				Description: "Repository ID",
				Required:    true,
			},
			"project_id": schema.Int64Attribute{
				Description: "Project ID",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}
