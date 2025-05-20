package group_users

import (
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var dataSourceSchema = datasource_schema.Schema{
	Description: "Databricks credential data source",
	Attributes: map[string]datasource_schema.Attribute{
		"id": datasource_schema.StringAttribute{
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
			Computed:    true,
		},
		"group_id": datasource_schema.Int64Attribute{
			Description: "ID of the group",
			Required:    true,
		},
		"users": datasource_schema.SetNestedAttribute{
			Computed:    true,
			Description: "List of users (map of ID and email) in the group",
			NestedObject: datasource_schema.NestedAttributeObject{
				Attributes: map[string]datasource_schema.Attribute{
					"id": datasource_schema.Int64Attribute{
						Description: "ID of the user",
						Required:    true,
					},
					"email": datasource_schema.StringAttribute{
						Description: "Email of the user",
						Required:    true,
					},
				},
			},
		},
	},
}
