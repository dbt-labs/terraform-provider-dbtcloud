package profile

import (
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var resourceSchema = resource_schema.Schema{
	Description: "Manages a dbt Cloud profile. A profile ties together a connection and credentials for use within environments.",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Description: "The ID of this resource. Contains the project ID and the profile ID.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"profile_id": resource_schema.Int64Attribute{
			Description: "The ID of the profile",
			Computed:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"project_id": resource_schema.Int64Attribute{
			Description: "The ID of the project in which to create the profile",
			Required:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"key": resource_schema.StringAttribute{
			Description: "Unique identifier for the profile",
			Required:    true,
		},
		"connection_id": resource_schema.Int64Attribute{
			Description: "The ID of the connection to use for this profile",
			Required:    true,
		},
		"credentials_id": resource_schema.Int64Attribute{
			Description: "The ID of the credentials to use for this profile",
			Required:    true,
		},
		"extended_attributes_id": resource_schema.Int64Attribute{
			Description: "The ID of the extended attributes for this profile. Set to null to unset.",
			Optional:    true,
		},
	},
}

var dataSourceSchema = datasource_schema.Schema{
	Description: "Retrieve data for a single profile",
	Attributes: map[string]datasource_schema.Attribute{
		"id": datasource_schema.StringAttribute{
			Description: "The ID of this resource. Contains the project ID and the profile ID.",
			Computed:    true,
		},
		"profile_id": datasource_schema.Int64Attribute{
			Description: "The ID of the profile",
			Required:    true,
		},
		"project_id": datasource_schema.Int64Attribute{
			Description: "The project ID to which the profile belongs",
			Required:    true,
		},
		"key": datasource_schema.StringAttribute{
			Description: "Unique identifier for the profile",
			Computed:    true,
		},
		"connection_id": datasource_schema.Int64Attribute{
			Description: "The ID of the connection used by this profile",
			Computed:    true,
		},
		"credentials_id": datasource_schema.Int64Attribute{
			Description: "The ID of the credentials used by this profile",
			Computed:    true,
		},
		"extended_attributes_id": datasource_schema.Int64Attribute{
			Description: "The ID of the extended attributes for this profile",
			Computed:    true,
		},
	},
}

var dataSourceAllSchema = datasource_schema.Schema{
	Description: "Retrieve data for multiple profiles",
	Attributes: map[string]datasource_schema.Attribute{
		"project_id": datasource_schema.Int64Attribute{
			Description: "The project ID to filter profiles for",
			Required:    true,
		},
		"profiles": datasource_schema.SetNestedAttribute{
			Description: "The list of profiles",
			Computed:    true,
			NestedObject: datasource_schema.NestedAttributeObject{
				Attributes: map[string]datasource_schema.Attribute{
					"id": datasource_schema.StringAttribute{
						Description: "The ID of this resource. Contains the project ID and the profile ID.",
						Computed:    true,
					},
					"profile_id": datasource_schema.Int64Attribute{
						Description: "The ID of the profile",
						Computed:    true,
					},
					"project_id": datasource_schema.Int64Attribute{
						Description: "The project ID to which the profile belongs",
						Computed:    true,
					},
					"key": datasource_schema.StringAttribute{
						Description: "Unique identifier for the profile",
						Computed:    true,
					},
					"connection_id": datasource_schema.Int64Attribute{
						Description: "The ID of the connection used by this profile",
						Computed:    true,
					},
					"credentials_id": datasource_schema.Int64Attribute{
						Description: "The ID of the credentials used by this profile",
						Computed:    true,
					},
					"extended_attributes_id": datasource_schema.Int64Attribute{
						Description: "The ID of the extended attributes for this profile",
						Computed:    true,
					},
				},
			},
		},
	},
}
