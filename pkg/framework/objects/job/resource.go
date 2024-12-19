package job

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = &jobResource{}
	_ resource.ResourceWithConfigure        = &jobResource{}
	_ resource.ResourceWithConfigValidators = &jobResource{}
	_ resource.ResourceWithImportState      = &jobResource{}
)

func JobResource() resource.Resource {
	return &jobResource{}
}

type jobResource struct {
	client *dbt_cloud.Client
}

func (d *jobResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

// Configure implements resource.ResourceWithConfigure.
func (d *jobResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	switch c := req.ProviderData.(type) {
	case nil: // do nothing
	case *dbt_cloud.Client:
		d.client = c
	default:
		resp.Diagnostics.AddError("Missing client", "A client is required to configure the job data source")
	}
}

// Schema implements resource.Resource.
func (d *jobResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the this resource",
				Computed:    true,
			},
			"project_id": schema.Int64Attribute{
				Description: "Project ID to create the job in",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.Int64Attribute{
				Description: "Environment ID to create the job in",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Job name",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Long Description for the job",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"execute_steps": schema.ListAttribute{
				Description: "List of commands to execute for the job",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"dbt_version": schema.StringAttribute{
				Description: "Version number of dbt to use in this job, usually in the format 1.2.0-latest rather than core versions",
				Optional:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Should always be set to true as setting it to false is the same as creating a job in a deleted state. To create/keep a job in a 'deactivated' state, check  the `triggers` config.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"triggers": schema.MapAttribute{ // TODO(cwalden) use SingleNestedAttribute?
				Description: "Flags for which types of triggers to use, the values are `github_webhook`, `git_provider_webhook`, `schedule` and `on_merge`. All flags should be listed and set with `true` or `false`. When `on_merge` is `true`, all the other values must be false.<br>`custom_branch_only` used to be allowed but has been deprecated from the API. The jobs will use the custom branch of the environment. Please remove the `custom_branch_only` from your config. <br>To create a job in a 'deactivated' state, set all to `false`.",
				Required:    true,
				ElementType: types.BoolType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.RequiresReplaceIfFuncResponse) {
							panic("unimplemented")
						},
						"",
						"",
					),
				},
			},
			// "triggers": schema.SingleNestedAttribute{
			// 	Description: "Flags for which types of triggers to use, keys of github_webhook, git_provider_webhook, schedule, on_merge",
			// 	Required:    true,
			// 	Attributes: map[string]schema.Attribute{
			// 		"github_webhook": schema.BoolAttribute{
			// 			Description: "Whether the job should be triggered by a GitHub webhook",
			// 			Optional:    true,
			// 			Default:     booldefault.StaticBool(false),
			// 		},
			// 		"git_provider_webhook": schema.BoolAttribute{
			// 			Description: "Whether the job should be triggered by a Git provider webhook",
			// 			Optional:    true,
			// 			Default:     booldefault.StaticBool(false),
			// 		},
			// 		"schedule": schema.BoolAttribute{
			// 			Description: "Whether the job should be triggered by a schedule",
			// 			Optional:    true,
			// 			Default:     booldefault.StaticBool(false),
			// 		},
			// 		"on_merge": schema.BoolAttribute{
			// 			Description: "Whether the job should be triggered by a merge",
			// 			Optional:    true,
			// 			Default:     booldefault.StaticBool(false),
			// 		},
			// 	},
			// 	PlanModifiers: []planmodifier.Object{
			// 		objectplanmodifier.RequiresReplaceIf(
			// 			func(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
			// 				panic("unimplemented")
			// 			},
			// 			"",
			// 			"",
			// 		),
			// 	},
			// },
			"num_threads": schema.Int64Attribute{
				Description: "Number of threads to use for the job",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
			},
			"target_name": schema.StringAttribute{
				Description: "Target name for the dbt profile",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("default"),
			},
			"generate_docs": schema.BoolAttribute{
				Description: "Flag for whether the job should generate documentation",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"run_generate_sources": schema.BoolAttribute{
				Description: "Flag for whether the job should add a `dbt source freshness` step to the job. The difference between manually adding a step with `dbt source freshness` in the job steps or using this flag is that with this flag, a failed freshness will still allow the following steps to run.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"schedule_type": schema.StringAttribute{
				Description: "Type of schedule to use, one of `every_day` / `days_of_week` / `custom_cron`",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("every_day"),
				Validators: []validator.String{
					stringvalidator.OneOf("every_day", "days_of_week", "custom_cron"),
				},
			},
			"schedule_interval": schema.Int64Attribute{
				Description: "Number of hours between job executions if running on a schedule",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 23),
				},
			},
			"schedule_hours": schema.SetAttribute{
				Description: "List of hours to execute the job at if running on a schedule",
				Optional:    true,
				ElementType: types.Int64Type,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueInt64sAre(
						int64validator.Between(1, 23),
					),
				},
			},
			"schedule_days": schema.SetAttribute{
				Description: "List of days of week as numbers (0 = Sunday, 7 = Saturday) to execute the job at if running on a schedule",
				Optional:    true,
				ElementType: types.Int64Type,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueInt64sAre(
						int64validator.Between(0, 7),
					),
				},
			},
			"schedule_cron": schema.StringAttribute{
				Description: "Custom `cron` expression to use for the schedule",
				Optional:    true,
				// TODO(cwalden) validate cron?
			},
			"deferring_job_id": schema.Int64Attribute{
				Description: "Job identifier that this job defers to (legacy deferring approach)",
				Optional:    true,
			},
			"deferring_environment_id": schema.Int64Attribute{
				Description: "Environment identifier that this job defers to (new deferring approach)",
				Optional:    true,
			},
			"self_deferring": schema.BoolAttribute{
				Description: "Whether this job defers on a previous run of itself",
				Optional:    true,
			},
			"timeout_seconds": schema.Int64Attribute{
				Description: "Number of seconds to allow the job to run before timing out",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"triggers_on_draft_pr": schema.BoolAttribute{
				Description: "Whether the CI job should be automatically triggered on draft PRs",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"run_compare_changes": schema.BoolAttribute{
				Description: "Whether the CI job should compare data changes introduced by the code changes. Requires `deferring_environment_id` to be set. (Advanced CI needs to be activated in the dbt Cloud Account Settings first as well)",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
		Blocks: map[string]schema.Block{
			"job_completion_trigger_condition": schema.SetNestedBlock{
				Description: "Which other job should trigger this job when it finishes, and on which conditions (sometimes referred as 'job chaining').",
				Validators: []validator.Set{
					setvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"job_id": schema.Int64Attribute{
							Description: "The ID of the job that would trigger this job after completion.",
							Required:    true,
						},
						"project_id": schema.Int64Attribute{
							Description: "The ID of the project where the trigger job is running in.",
							Required:    true,
						},
						"statuses": schema.SetAttribute{
							Description: "List of statuses to trigger the job on. Possible values are `success`, `error` and `canceled`.",
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// ConfigValidators implements resource.ResourceWithConfigValidators.
func (d *jobResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("schedule_interval"),
			path.MatchRoot("schedule_hours"),
			path.MatchRoot("schedule_cron"),
		),
		resourcevalidator.Conflicting(
			path.MatchRoot("self_deferring"),
			path.MatchRoot("deferring_job_id"),
			path.MatchRoot("deferring_environment_id"),
		),
	}
}

// Read implements resource.Resource.
func (d *jobResource) Read(context.Context, resource.ReadRequest, *resource.ReadResponse) {
	panic("unimplemented")
}

// Create implements resource.Resource.
func (d *jobResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
	panic("unimplemented")
}

// Update implements resource.Resource.
func (d *jobResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	panic("unimplemented")
}

// Delete implements resource.Resource.
func (d *jobResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	panic("unimplemented")
}

// ImportState implements resource.ResourceWithImportState.
func (d *jobResource) ImportState(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse) {
	panic("unimplemented")
}
