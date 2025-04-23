package environment_variable

import (
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var resourceSchema = resource_schema.Schema{
	Description: "Environment variable resource",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource. Contains the project ID and the environment variable ID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": resource_schema.Int64Attribute{
			Required:    true,
			Description: "Project ID to create the environment variable in",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"name": resource_schema.StringAttribute{
			Required:    true,
			Description: "Name for the variable, must be unique within a project, must be prefixed with 'DBT_'",
		},
		"environment_values": resource_schema.MapAttribute{
			Required:    true,
			ElementType: types.StringType,
			Description: "Map from environment names to respective variable value, a special key `project` should be set for the project default variable value. This field is not set as sensitive so take precautions when using secret environment variables.",
		},
	},
}

var datasourceSchema = datasource_schema.Schema{
	Description: "Environment variable credential data source",
	Attributes: map[string]datasource_schema.Attribute{
		"id": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource. Contains the project ID and the environment variable ID.",
		},
		"project_id": datasource_schema.Int64Attribute{
			Required:    true,
			Description: "Project ID to create the environment variable in",
		},
		"name": datasource_schema.StringAttribute{
			Required:    true,
			Description: "Name for the variable, must be unique within a project, must be prefixed with 'DBT_'",
		},
		"environment_values": datasource_schema.MapAttribute{
			Computed:    true,
			ElementType: types.StringType,
			Description: "Map from environment names to respective variable value, a special key `project` should be set for the project default variable value. This field is not set as sensitive so take precautions when using secret environment variables.",
		},
	},
}
