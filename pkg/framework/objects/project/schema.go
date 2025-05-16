package project

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (d *projectsDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve all the projects created in dbt Cloud with an optional filter on parts of the project name.",
		Attributes: map[string]schema.Attribute{
			"name_contains": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Used to filter projects by name, Optional",
			},
			"projects": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Set of projects with their details",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "Project ID",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Project name",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Project description",
						},
						"semantic_layer_config_id": schema.Int64Attribute{
							Computed:    true,
							Description: "Semantic layer config ID",
						},
						"dbt_project_subdirectory": schema.StringAttribute{
							Computed:    true,
							Description: "Subdirectory for the dbt project inside the git repo",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "When the project was created",
						},
						"updated_at": schema.StringAttribute{
							Computed:    true,
							Description: "When the project was last updated",
						},
						"repository": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Details for the repository linked to the project",
							Attributes: map[string]schema.Attribute{
								"id": schema.Int64Attribute{
									Computed:    true,
									Description: "Repository ID",
								},
								"remote_url": schema.StringAttribute{
									Computed:    true,
									Description: "URL of the git repo remote",
								},
								"pull_request_url_template": schema.StringAttribute{
									Computed:    true,
									Description: "URL template for PRs",
								},
							},
						},
						"connection": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Details for the connection linked to the project",
							Attributes: map[string]schema.Attribute{
								"id": schema.Int64Attribute{
									Computed:    true,
									Description: "Connection ID",
								},
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "Connection name",
								},
								"adapter_version": schema.StringAttribute{
									Computed:    true,
									Description: "Version of the adapter for the connection. Will tell what connection type it is",
								},
							},
						},
					},
				},
			},
		},
	}
}
