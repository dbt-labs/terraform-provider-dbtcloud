package environment_variable_job_override

import (
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var resourceSchema = resource_schema.Schema{
	Description: "Environment variable job override resource",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource. Contains the project ID and the environment variable job override ID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"environment_variable_job_override_id": resource_schema.Int64Attribute{
			Computed:    true,
			Description: "The internal ID of this resource. Contains the project ID and the environment variable job override ID.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"job_definition_id": resource_schema.Int64Attribute{
			Required:    true,
			Description: "The job ID for which the environment variable is being overridden",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"project_id": resource_schema.Int64Attribute{
			Required:    true,
			Description: "Project ID to create the environment variable job override in",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"raw_value": resource_schema.StringAttribute{
			Required:    true,
			Description: "The value for the override of the environment variable",
		},
		"name": resource_schema.StringAttribute{
			Required:    true,
			Description: "The environment variable name to override",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	},
}
