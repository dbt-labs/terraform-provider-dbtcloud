package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (d *userDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve user details",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the user",
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "Email for the user",
			},
		},
	}
}

func (d *usersDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve all users",
		Attributes: map[string]schema.Attribute{
			"users": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Set of users with their internal ID end email",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "ID of the user",
						},
						"email": schema.StringAttribute{
							Computed:    true,
							Description: "Email for the user",
						},
					},
				},
			},
		},
	}
}
