package job

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &jobDataSource{}
	_ datasource.DataSourceWithConfigure = &jobDataSource{}
)

func JobDataSource() datasource.DataSource {
	return &jobDataSource{}
}

type jobDataSource struct {
	client *dbt_cloud.Client
}

func (d *jobDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *jobDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	switch c := req.ProviderData.(type) {
	case nil: // do nothing
	case *dbt_cloud.Client:
		d.client = c
	default:
		resp.Diagnostics.AddError("Missing client", "A client is required to configure the job data source")
	}
}

// Schema implements datasource.DataSourceWithValidateConfig.
func (d *jobDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"job_id": schema.Int64Attribute{
				Description: "ID of the job",
				Required:    true,
			},
			"project_id": schema.Int64Attribute{
				Description: "ID of the project the job is in",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the this resource",
				Computed:    true,
			},
			"environment_id": schema.Int64Attribute{
				Description: "ID of the environment the job is in",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Given name for the job",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Long description for the job",
				Computed:    true,
			},
			"deferring_job_id": schema.Int64Attribute{
				Description: "ID of the job this job defers to",
				Computed:    true,
			},
			"deferring_environment_id": schema.Int64Attribute{
				Description: "ID of the environment this job defers to",
				Computed:    true,
			},
			"self_deferring": schema.BoolAttribute{
				Description: "Whether this job defers on a previous run of itself (overrides value in deferring_job_id)",
				Computed:    true,
			},
			"triggers": schema.MapAttribute{
				Description: "Flags for which types of triggers to use, keys of github_webhook, git_provider_webhook, schedule, on_merge",
				Computed:    true,
				ElementType: types.BoolType,
			},
			"timeout_seconds": schema.Int64Attribute{
				Description: "Number of seconds before the job times out",
				Computed:    true,
			},
			"triggers_on_draft_pr": schema.BoolAttribute{
				Description: "Whether the CI job should be automatically triggered on draft PRs",
				Computed:    true,
			},
			// "job_completion_trigger_condition": schema.NestedSingleAttribute{

			"run_compare_changes": schema.BoolAttribute{
				Description: "Whether the CI job should compare data changes introduced by the code change in the PR.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"job_completion_trigger_condition": schema.SetNestedBlock{
				Description: "Whether the CI job should compare data changes introduced by the code change in the PR.",
				Validators: []validator.Set{
					setvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"job_id": schema.Int64Attribute{
							Description: "The ID of the job that would trigger this job after completion.",
							Computed:    true,
						},
						"project_id": schema.Int64Attribute{
							Description: "The ID of the project where the trigger job is running in.",
							Computed:    true,
						},
						"statuses": schema.SetAttribute{
							Description: "List of statuses to trigger the job on.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Read implements datasource.DataSourceWithValidateConfig.
func (d *jobDataSource) Read(context.Context, datasource.ReadRequest, *datasource.ReadResponse) {
	panic("unimplemented")
}
