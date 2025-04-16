package project

import (
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var datasourceSchema = datasource_schema.Schema{
	Description: "Retrieve all the projects created in dbt Cloud with an optional filter on parts of the project name.",
	Attributes: map[string]datasource_schema.Attribute{
		"name_contains": datasource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Used to filter projects by name, Optional",
		},
		"projects": datasource_schema.SetNestedAttribute{
			Computed:    true,
			Description: "Set of projects with their details",
			NestedObject: datasource_schema.NestedAttributeObject{
				Attributes: map[string]datasource_schema.Attribute{
					"id": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "Project ID",
					},
					"name": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Project name",
					},
					"description": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Project description",
					},
					"semantic_layer_config_id": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "Semantic layer config ID",
					},
					"dbt_project_subdirectory": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Subdirectory for the dbt project inside the git repo",
					},
					"created_at": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "When the project was created",
					},
					"updated_at": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "When the project was last updated",
					},
					"repository": datasource_schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Details for the repository linked to the project",
						Attributes: map[string]datasource_schema.Attribute{
							"id": datasource_schema.Int64Attribute{
								Computed:    true,
								Description: "Repository ID",
							},
							"remote_url": datasource_schema.StringAttribute{
								Computed:    true,
								Description: "URL of the git repo remote",
							},
							"pull_request_url_template": datasource_schema.StringAttribute{
								Computed:    true,
								Description: "URL template for PRs",
							},
						},
					},
					"project_connection": datasource_schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Details for the connection linked to the project",
						Attributes: map[string]datasource_schema.Attribute{
							"id": datasource_schema.Int64Attribute{
								Computed:    true,
								Description: "Connection ID",
							},
							"name": datasource_schema.StringAttribute{
								Computed:    true,
								Description: "Connection name",
							},
							"adapter_version": datasource_schema.StringAttribute{
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

var singleDatasourceSchema = datasource_schema.Schema{
	Description: "Retrieve a specific project from dbt Cloud.",
	Attributes: map[string]datasource_schema.Attribute{
		"id": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: "Project ID",
		},
		"name": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "Project name",
		},
		"description": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "Project description",
		},
		"semantic_layer_config_id": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: "Semantic layer config ID",
		},
		"dbt_project_subdirectory": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "Subdirectory for the dbt project inside the git repo",
		},
		"created_at": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "When the project was created",
		},
		"updated_at": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "When the project was last updated",
		},
		"repository": datasource_schema.SingleNestedAttribute{
			Computed:    true,
			Description: "Details for the repository linked to the project",
			Attributes: map[string]datasource_schema.Attribute{
				"id": datasource_schema.Int64Attribute{
					Computed:    true,
					Description: "Repository ID",
				},
				"remote_url": datasource_schema.StringAttribute{
					Computed:    true,
					Description: "URL of the git repo remote",
				},
				"pull_request_url_template": datasource_schema.StringAttribute{
					Computed:    true,
					Description: "URL template for PRs",
				},
			},
		},
		"project_connection": datasource_schema.SingleNestedAttribute{
			Computed:    true,
			Description: "Details for the connection linked to the project",
			Attributes: map[string]datasource_schema.Attribute{
				"id": datasource_schema.Int64Attribute{
					Computed:    true,
					Description: "Connection ID",
				},
				"name": datasource_schema.StringAttribute{
					Computed:    true,
					Description: "Connection name",
				},
				"adapter_version": datasource_schema.StringAttribute{
					Computed:    true,
					Description: "Version of the adapter for the connection. Will tell what connection type it is",
				},
			},
		},
		"freshness_job_id": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: "ID of Job for source freshness",
		},
		"docs_job_id": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: "ID of Job for the documentation",
		},
		"state": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: "Project state should be 1 = active, as 2 = deleted",
		},
	},
}

var resourceSchema = resource_schema.Schema{
	Description: "Manages a dbt Cloud project.",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.Int64Attribute{
			Description: "The ID of the project.",
			Computed:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": resource_schema.StringAttribute{
			Description: "Project name",
			Required:    true,
		},
		"description": resource_schema.StringAttribute{
			Description: "Description for the project. Will show in dbt Explorer.",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"dbt_project_subdirectory": resource_schema.StringAttribute{
			Description: "DBT project subdirectory",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	},
}
