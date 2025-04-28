package runs

import (
	all_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var allDatasourceSchema = all_schema.Schema{
	Description: "Retrieve all runs",
	Attributes: map[string]all_schema.Attribute{
		"filter": all_schema.SingleNestedAttribute{
			Optional:    true,
			Description: "Filter to apply to the runs",
			Attributes: map[string]all_schema.Attribute{
				"environment_id": all_schema.Int64Attribute{
					Optional:    true,
					Description: "The ID of the environment",
				},
				"limit": all_schema.Int64Attribute{
					Optional:    true,
					Description: "The limit of the runs",
				},
				"project_id": all_schema.Int64Attribute{
					Optional:    true,
					Description: "The ID of the project",
				},
				"trigger_id": all_schema.Int64Attribute{
					Optional:    true,
					Description: "The ID of the trigger",
				},
				"job_definition_id": all_schema.Int64Attribute{
					Optional:    true,
					Description: "The ID of the job definition",
				},
				"pull_request_id": all_schema.Int64Attribute{
					Optional:    true,
					Description: "The ID of the pull request",
				},
				"status": all_schema.Int64Attribute{
					Optional:    true,
					Description: "The status of the run",
				},
				"status_in": all_schema.StringAttribute{
					Optional:    true,
					Description: "The status of the run",
				},
			},
		},
		"runs": all_schema.SetNestedAttribute{
			Computed:    true,
			Description: "Set of users with their internal ID end email",
			NestedObject: all_schema.NestedAttributeObject{
				Attributes: map[string]all_schema.Attribute{
					"id": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The ID of the run",
					},
					"account_id": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The ID of the account",
					},
					"job_id": datasource_schema.Int64Attribute{
						Required:    true,
						Description: "The ID of the job",
					},
					"git_sha": datasource_schema.StringAttribute{
						Required:    true,
						Description: "The SHA of the commit",
					},
					"git_branch": datasource_schema.StringAttribute{
						Required:    true,
						Description: "The branch of the commit",
					},
					"github_pull_request_id": datasource_schema.StringAttribute{
						Required:    true,
						Description: "The ID of the pull request",
					},
					"schema_override": datasource_schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The schema override",
					},
					"cause": datasource_schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The cause of the run",
					},
				},
			},
		},
	},
}

var datasourceSchema = datasource_schema.Schema{
	Description: "Retrieve all the runs created in dbt Cloud",
	Attributes: map[string]datasource_schema.Attribute{
		"id": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the run",
		},
		"account_id": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the account",
		},
		"job_id": datasource_schema.Int64Attribute{
			Required:    true,
			Description: "The ID of the job",
		},
		"git_sha": datasource_schema.StringAttribute{
			Required:    true,
			Description: "The SHA of the commit",
		},
		"git_branch": datasource_schema.StringAttribute{
			Required:    true,
			Description: "The branch of the commit",
		},
		"github_pull_request_id": datasource_schema.StringAttribute{
			Required:    true,
			Description: "The ID of the pull request",
		},
		"schema_override": datasource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The schema override",
		},
		"cause": datasource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The cause of the run",
		},
	},
}
