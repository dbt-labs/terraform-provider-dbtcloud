package job

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getJobAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"execution": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"timeout_seconds": schema.Int64Attribute{
					Computed:    true,
					Description: "The number of seconds before the job times out",
				},
			},
		},
		"timeout_seconds": schema.Int64Attribute{
			Computed:    true,
			Description: "[Deprecated - Moved to execution.timeout_seconds] Number of seconds before the job times out",
		},
		"generate_docs": schema.BoolAttribute{
			Computed:    true,
			Description: "Whether the job generate docs",
		},
		"run_generate_sources": schema.BoolAttribute{
			Computed:    true,
			Description: "Whether the job test source freshness",
		},
		"run_compare_changes": schema.BoolAttribute{
			Computed:    true,
			Description: "Whether the job should compare data changes introduced by the code change in the PR",
		},
		"id": schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the job",
		},
		"project_id": schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the project",
		},
		"environment_id": schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of environment",
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: "The name of the job",
		},
		"description": schema.StringAttribute{
			Computed:    true,
			Description: "The description of the job",
		},
		"dbt_version": schema.StringAttribute{
			Computed:    true,
			Description: "The version of dbt used for the job. If not set, the environment version will be used.",
		},
		"execute_steps": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
			Description: "The list of steps to run in the job",
		},
		"deferring_job_id": schema.Int64Attribute{
			Computed:    true,
			Description: "[Deprecated - deferral is now set at the environment level] The ID of the job definition this job defers to",
		},
		"deferring_job_definition_id": schema.Int64Attribute{
			Computed:    true,
			Description: "[Deprecated - deferral is now set at the environment level] The ID of the job definition this job defers to",
		},
		"self_deferring": schema.BoolAttribute{
			Computed:    true,
			Description: "Whether this job defers on a previous run of itself (overrides value in deferring_job_id)",
		},
		"deferring_environment_id": schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the environment this job defers to",
		},
		"triggers": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"github_webhook": schema.BoolAttribute{
					Computed:    true,
					Description: "Whether the job runs automatically on PR creation",
				},
				"git_provider_webhook": schema.BoolAttribute{
					Computed:    true,
					Description: "Whether the job runs automatically on PR creation",
				},
				"schedule": schema.BoolAttribute{
					Computed:    true,
					Description: "Whether the job runs on a schedule",
				},
				"on_merge": schema.BoolAttribute{
					Computed:    true,
					Description: "Whether the job runs automatically once a PR is merged",
				},
			},
		},
		// todo - decide what to do with this
		// "triggers_old": schema.MapAttribute{
		// 	Computed:    true,
		// 	ElementType: types.BoolType,
		// 	Description: "Flags for which types of triggers to use, keys of github_webhook, git_provider_webhook, schedule, on_merge",
		// },
		"settings": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"threads": schema.Int64Attribute{
					Computed:    true,
					Description: "Number of threads to run dbt with",
				},
				"target_name": schema.StringAttribute{
					Computed:    true,
					Description: "Value for `target.name` in the Jinja context",
				},
			},
		},
		"schedule": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"cron": schema.StringAttribute{
					Computed:    true,
					Description: "The cron schedule for the job. Only used if triggers.schedule is true",
				},
			},
		},
		"job_type": schema.StringAttribute{
			Computed:    true,
			Description: "The type of job (e.g. CI, scheduled)",
		},
		"triggers_on_draft_pr": schema.BoolAttribute{
			Computed:    true,
			Description: "Whether the CI job should be automatically triggered on draft PRs",
		},
		"environment": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "Details of the environment the job is running in",
			Attributes: map[string]schema.Attribute{
				"project_id": schema.Int64Attribute{
					Computed: true,
				},
				"id": schema.Int64Attribute{
					Computed:    true,
					Description: "ID of the environment",
				},
				"name": schema.StringAttribute{
					Computed:    true,
					Description: "Name of the environment",
				},
				"deployment_type": schema.StringAttribute{
					Computed:    true,
					Description: "Type of deployment environment: staging, production",
				},
				"type": schema.StringAttribute{
					Computed:    true,
					Description: "Environment type: development or deployment",
				},
			},
		},
		"job_completion_trigger_condition": schema.SingleNestedAttribute{
			Computed:    true,
			Optional: true,
			Description: "Whether the job is triggered by the completion of another job",
			Attributes: map[string]schema.Attribute{
				"condition": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"job_id": schema.Int64Attribute{
							Computed: true,
						},
						"project_id": schema.Int64Attribute{
							Computed: true,
						},
						"statuses": schema.SetAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d jobsDataSource) ValidateConfig(
	ctx context.Context,
	req datasource.ValidateConfigRequest,
	resp *datasource.ValidateConfigResponse,
) {
	var data JobsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ProjectID.IsNull() && data.EnvironmentID.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("project_id"),
			"Missing Attribute Configuration",
			"project_id or environment_id must be configured.",
		)
	}

	if !(data.ProjectID.IsNull() || data.EnvironmentID.IsNull()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("project_id"),
			"Invalid Attribute Configuration",
			"Only one of project_id or environment_id can be configured.",
		)
	}
}

func (j *jobDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	jobAttributes := getJobAttributes()
	
	jobAttributes["job_id"] = schema.Int64Attribute{
		Required: true,
		Description: "The ID of the job",
	}
	
	resp.Schema = schema.Schema{
		Description: "Get detailed information for a specific dbt Cloud job.",
		Attributes: jobAttributes,
	}
}

func (d *jobsDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {

	jobAttributes := getJobAttributes()
	jobAttributes["job_id"] = schema.Int64Attribute{
		Computed: true,
		Description: "The ID of the job",
	}
	resp.Schema = schema.Schema{
		Description: "Retrieve all the jobs for a given dbt Cloud project or environment along with the environment details for the jobs. This will return both the jobs created from Terraform but also the jobs created in the dbt Cloud UI.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.Int64Attribute{
				Optional:    true,
				Description: "The ID of the project for which we want to retrieve the jobs (one of `project_id` or `environment_id` must be set)",
			},
			"environment_id": schema.Int64Attribute{
				Optional:    true,
				Description: "The ID of the environment for which we want to retrieve the jobs (one of `project_id` or `environment_id` must be set)",
			},
			"jobs": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Set of jobs with their details",
				NestedObject: schema.NestedAttributeObject{
					Attributes: jobAttributes,
				},
			},
		},
	}
}


