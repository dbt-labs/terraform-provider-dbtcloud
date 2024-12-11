package azure_dev_ops_repository

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (d *azureDevOpsRepositoryDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: `Use this data source to retrieve the ID and details of an Azure Dev Ops repository 
based on its name and the ID of the Azure Dev Ops project it belongs to.
		
This data source requires connecting with a user token and doesn't work with a service token.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The internal Azure Dev Ops ID of the ADO Repository",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the ADO repository",
			},
			"azure_dev_ops_project_id": schema.StringAttribute{
				Required:    true,
				Description: "The internal Azure Dev Ops ID of the ADO Project. Can be retrieved using the data source dbtcloud_azure_dev_ops_project and the project name",
			},
			"details_url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the ADO repository showing details about the repository and its attributes",
			},
			"remote_url": schema.StringAttribute{
				Computed:    true,
				Description: "The HTTP URL of the ADO repository used to connect to dbt Cloud",
			},
			"web_url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the ADO repository accessible in the browser",
			},
			"default_branch": schema.StringAttribute{
				Computed:    true,
				Description: "The default branch of the ADO repository",
			},
		},
	}

}
