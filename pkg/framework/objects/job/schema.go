package job

import (
	"context"

	job_validators "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/job/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
			Computed:           true,
			DeprecationMessage: "Moved to execution.timeout_seconds",
			Description:        "[Deprectated - Moved to execution.timeout_seconds] Number of seconds before the job times out",
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
		"deferring_environment_id": schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the environment this job defers to",
		},
		"force_node_selection": schema.BoolAttribute{
			Computed:    true,
			Description: "Whether force node selection (SAO) is enabled for this job",
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
		Required:    true,
		Description: "The ID of the job",
	}

	jobAttributes["deferring_job_id"] = schema.Int64Attribute{
		Computed:           true,
		DeprecationMessage: "Deferral is now set at the environment level",
		Description:        "[Deprectated - Deferral is now set at the environment level] The ID of the job definition this job defers to",
	}

	jobAttributes["self_deferring"] = schema.BoolAttribute{
		Computed:    true,
		Description: "Whether this job defers on a previous run of itself (overrides value in deferring_job_id)",
	}

	jobAttributes["job_completion_trigger_condition"] = schema.ListNestedAttribute{
		Computed:    true,
		Optional:    true,
		Description: "Which other job should trigger this job when it finishes, and on which conditions. Format for the property will change in the next release to match the one from the one from dbtcloud_jobs.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"job_id": schema.Int64Attribute{
					Computed:    true,
					Description: "The ID of the job that would trigger this job after completion.",
				},
				"project_id": schema.Int64Attribute{
					Computed:    true,
					Description: "The ID of the project where the trigger job is running in.",
				},
				"statuses": schema.SetAttribute{
					Computed:    true,
					ElementType: types.StringType,
					Description: "List of statuses to trigger the job on.",
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Get detailed information for a specific dbt Cloud job.",
		Attributes:  jobAttributes,
	}
}

func (d *jobsDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {

	jobAttributes := getJobAttributes()
	jobAttributes["job_id"] = schema.Int64Attribute{
		Computed:    true,
		Description: "The ID of the job",
	}
	jobAttributes["deferring_job_definition_id"] = schema.Int64Attribute{
		Computed:           true,
		DeprecationMessage: "Deferral is now set at the environment level",
		Description:        "[Deprectated - Deferral is now set at the environment level] The ID of the job definition this job defers to",
	}
	jobAttributes["job_completion_trigger_condition"] = schema.SingleNestedAttribute{
		Computed:    true,
		Optional:    true,
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

var (
	scheduleTypes = []string{
		"every_day",
		"days_of_week",
		"custom_cron",
		"interval_cron",
	}
)

func (j *jobResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resource_schema.Schema{
		Description: "Managed a dbt Cloud job.",
		Attributes: map[string]resource_schema.Attribute{
			"id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of this resource",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			// "execution": resource_schema.SingleNestedAttribute{
			// 	Optional: true,
			// 	Computed: true,
			// 	Attributes: map[string]resource_schema.Attribute{
			// 		"timeout_seconds": resource_schema.Int64Attribute{
			// 			Optional:    true,
			// 			Computed:    true,
			// 			Default:     int64default.StaticInt64(0),
			// 			Description: "The number of seconds before the job times out",
			// 		},
			// 	},
			// },
			"timeout_seconds": resource_schema.Int64Attribute{
				Optional:           true,
				Computed:           true,
				Default:            int64default.StaticInt64(0),
				DeprecationMessage: "Moved to execution.timeout_seconds",
				Description:        "[Deprectated - Moved to execution.timeout_seconds] Number of seconds to allow the job to run before timing out",
			},
			"generate_docs": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Flag for whether the job should generate documentation",
			},
			"run_generate_sources": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Flag for whether the job should add a `dbt source freshness` step to the job. The difference between manually adding a step with `dbt source freshness` in the job steps or using this flag is that with this flag, a failed freshness will still allow the following steps to run.",
			},
			"run_lint": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the CI job should lint SQL changes. Defaults to `false`.",
			},
			"errors_on_lint_failure": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the CI job should fail when a lint error is found. Only used when `run_lint` is set to `true`. Defaults to `true`.",
			},
			"schedule_type": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("every_day"),
				Description: "Type of schedule to use, one of every_day/ days_of_week/ custom_cron/ interval_cron",
				Validators: []validator.String{
					stringvalidator.OneOf(scheduleTypes...),
				},
			},
			"schedule_interval": resource_schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Description: "Number of hours between job executions if running on a schedule",
				Validators: []validator.Int64{
					int64validator.Between(1, 23),
					int64validator.ConflictsWith(
						path.MatchRoot("schedule_hours"),
						path.MatchRoot("schedule_cron"),
					),
				},
			},
			"schedule_hours": resource_schema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
				Description: "List of hours to execute the job at if running on a schedule",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ConflictsWith(
						path.MatchRoot("schedule_interval"),
						path.MatchRoot("schedule_cron"),
					),
				},
			},
			"schedule_days": resource_schema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
				Description: "List of days of week as numbers (0 = Sunday, 7 = Saturday) to execute the job at if running on a schedule",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"schedule_cron": resource_schema.StringAttribute{
				Optional:    true,
				Description: "Custom cron expression for schedule",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("schedule_interval"),
						path.MatchRoot("schedule_hours"),
					),
				},
			},
			"run_compare_changes": resource_schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				// TODO Once on the plugin framework, put a validation to check that `deferring_environment_id` is set
				Description: "Whether the CI job should compare data changes introduced by the code changes. Requires `deferring_environment_id` to be set. (Advanced CI needs to be activated in the dbt Cloud Account Settings first as well)",
			},
			"job_id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "Job identifier",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"project_id": resource_schema.Int64Attribute{
				Required:    true,
				Description: "Project ID to create the job in",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"environment_id": resource_schema.Int64Attribute{
				Required:    true,
				Description: "Environment ID to create the job in",
			},
			"name": resource_schema.StringAttribute{
				Required:    true,
				Description: "Job name",
			},
			"description": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Description for the job",
			},
			"dbt_version": resource_schema.StringAttribute{
				Optional:    true,
				Description: "Version number of dbt to use in this job, usually in the format 1.2.0-latest rather than core versions",
			},
		"force_node_selection": resource_schema.BoolAttribute{
			Optional:           true,
			Computed:           true,
			DeprecationMessage: "Use cost_optimization_features instead. force_node_selection will be removed in a future version.",
			Description:        "Whether to force node selection (SAO - Select All Optimizations) for the job. If `dbt_version` is not set to `latest-fusion`, this must be set to `true` when specified. Deprecated: Use cost_optimization_features instead.",
			Validators: []validator.Bool{
				job_validators.ForceNodeSelectionValidator(),
			},
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"cost_optimization_features": resource_schema.SetAttribute{
			Optional:    true,
			Computed:    true,
			ElementType: types.StringType,
			Description: "List of cost optimization features enabled for the job. Valid values: `state_aware_orchestration`. When `state_aware_orchestration` is included, SAO is enabled (equivalent to force_node_selection=false). When empty or not set, SAO is disabled (equivalent to force_node_selection=true). This is the preferred way to control SAO; use this instead of force_node_selection.",
		},
			"execute_steps": resource_schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of commands to execute for the job",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"validate_execute_steps": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When set to `true`, the provider will validate the `execute_steps` during plan time to ensure they contain valid dbt commands. If a command is not recognized (e.g., a new dbt command not yet supported by the provider), the validation will fail. Defaults to `false` to allow flexibility with newer dbt commands.",
			},
			"is_active": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Should always be set to true as setting it to false is the same as creating a job in a deleted state. To create/keep a job in a 'deactivated' state, check  the `triggers` config. Setting it to false essentially deletes the job. On resource creation, this field is enforced to be true.",
			},
			"triggers": resource_schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]resource_schema.Attribute{
					"github_webhook": resource_schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether the job runs automatically on PR creation",
					},
					"git_provider_webhook": resource_schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether the job runs automatically on PR creation",
					},
					"schedule": resource_schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether the job runs on a schedule",
					},
					"on_merge": resource_schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether the job runs automatically once a PR is merged",
					},
				},
				Description: "Flags for which types of triggers to use, the values are `github_webhook`, `git_provider_webhook`, `schedule` and `on_merge`. All flags should be listed and set with `true` or `false`. When `on_merge` is `true`, all the other values must be false.<br>`custom_branch_only` used to be allowed but has been deprecated from the API. The jobs will use the custom branch of the environment. Please remove the `custom_branch_only` from your config. <br>To create a job in a 'deactivated' state, set all to `false`.",
			},
			"num_threads": resource_schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1),
				// todo mark deprecated
				Description: "Number of threads to use in the job",
			},
			"target_name": resource_schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("default"),
				// todo mark deprecated
				Description: "Target name for the dbt profile",
			},
			// "settings": resource_schema.SingleNestedAttribute{
			// 	Computed: true,
			// 	Attributes: map[string]resource_schema.Attribute{
			// 		"threads": resource_schema.Int64Attribute{
			// 			Optional:    true,
			// 			Computed:    true,
			// 			Default:     int64default.StaticInt64(1),
			// 			Description: "Number of threads to run dbt with",
			// 		},
			// 		"target_name": resource_schema.StringAttribute{
			// 			Optional:    true,
			// 			Computed:    true,
			// 			Default:     stringdefault.StaticString("default"),
			// 			Description: "Value for `target.name` in the Jinja context",
			// 		},
			// 	},
			// },
			"deferring_job_id": resource_schema.Int64Attribute{
				Optional:    true,
				Description: "Job identifier that this job defers to (legacy deferring approach)",
				Validators: []validator.Int64{
					int64validator.ConflictsWith(
						path.MatchRoot("self_deferring"),
						path.MatchRoot("deferring_environment_id"),
					),
				},
			},
			"deferring_environment_id": resource_schema.Int64Attribute{
				Optional:    true,
				Description: "Environment identifier that this job defers to (new deferring approach)",
				Validators: []validator.Int64{
					int64validator.ConflictsWith(
						path.MatchRoot("self_deferring"),
						path.MatchRoot("deferring_job_id"),
					),
				},
			},
			"self_deferring": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: " Whether this job defers on a previous run of itself",
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(
						path.MatchRoot("deferring_job_id"),
					),
				},
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"triggers_on_draft_pr": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the CI job should be automatically triggered on draft PRs",
			},
			"compare_changes_flags": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				// No default - only set when run_compare_changes is true
				// Setting a default causes SAO validation errors for CI/Merge jobs
				Description: "The model selector for checking changes in the compare changes Advanced CI feature",
			},
			"job_type": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Can be used to enforce the job type betwen `ci`, `merge` and `scheduled`. Without this value the job type is inferred from the triggers configured",
			},

			// todo add these after
			// "schedule": resource_schema.SingleNestedAttribute{
			// 	Computed: true,
			// 	Attributes: map[string]resource_schema.Attribute{
			// 		"cron": resource_schema.StringAttribute{
			// 			Computed:    true,
			// 			Description: "The cron schedule for the job. Only used if triggers.schedule is true",
			// 		},
			// 	},
			// },
			// "environment": resource_schema.SingleNestedAttribute{
			// 	Computed:    true,
			// 	Description: "Details of the environment the job is running in",
			// 	Attributes: map[string]resource_schema.Attribute{
			// 		"project_id": resource_schema.Int64Attribute{
			// 			Computed: true,
			// 		},
			// 		"id": resource_schema.Int64Attribute{
			// 			Computed:    true,
			// 			Description: "ID of the environment",
			// 		},
			// 		"name": resource_schema.StringAttribute{
			// 			Computed:    true,
			// 			Description: "Name of the environment",
			// 		},
			// 		"deployment_type": resource_schema.StringAttribute{
			// 			Computed:    true,
			// 			Description: "Type of deployment environment: staging, production",
			// 		},
			// 		"type": resource_schema.StringAttribute{
			// 			Computed:    true,
			// 			Description: "Environment type: development or deployment",
			// 		},
			// 	},
			// },
		},
		Blocks: map[string]resource_schema.Block{
			"job_completion_trigger_condition": resource_schema.ListNestedBlock{
				Description: "Which other job should trigger this job when it finishes, and on which conditions (sometimes referred as 'job chaining').",
				NestedObject: resource_schema.NestedBlockObject{
					Attributes: map[string]resource_schema.Attribute{
						"job_id": resource_schema.Int64Attribute{
							Required:    true,
							Description: "The ID of the job that would trigger this job after completion.",
						},
						"project_id": resource_schema.Int64Attribute{
							Required:    true,
							Description: "The ID of the project where the trigger job is running in.",
						},
						"statuses": resource_schema.SetAttribute{
							Required:    true,
							ElementType: types.StringType,
							Description: "List of statuses to trigger the job on. Possible values are `success`, `error` and `canceled`.",
						},
					},
				},
			},
		},
	}
}
