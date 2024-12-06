package azure_dev_ops_project

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (r *azureDevOpsProjectDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: `Use this data source to retrieve the ID of an Azure Dev Ops project 
based on its name.
		
This data source requires connecting with a user token and doesn't work with a service token.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The internal Azure Dev Ops ID of the ADO Project",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the ADO project",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the ADO project",
			},
		},
	}
}
