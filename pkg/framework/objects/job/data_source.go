package job

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

var (
	_ datasource.DataSource                   = &jobDataSource{}
	_ datasource.DataSourceWithConfigure      = &jobDataSource{}
	_ datasource.DataSourceWithValidateConfig = &jobDataSource{}
)

func JobDataSource() datasource.DataSource {
	return &jobDataSource{}
}

type jobDataSource struct {
	client *dbt_cloud.Client
}

func (j *jobDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Debug(context.Background(), "Configuring dbt Cloud job data source")
	
	if req.ProviderData == nil {
		return
	}

	j.client = req.ProviderData.(*dbt_cloud.Client)
	tflog.Debug(context.Background(), "Configured dbt Cloud job data source")
}

func (j *jobDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

func (j *jobDataSource) ValidateConfig(
	ctx context.Context,
	req datasource.ValidateConfigRequest,
	resp *datasource.ValidateConfigResponse,
) {
	var data JobDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.JobId.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("job_id"),
			"Missing Required Attribute",
			"job_id must be configured.",
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

func (j *jobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading dbt Cloud job data source - this should appear in logs")
	
	var state JobDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	
	// Print the entire state for debugging
	tflog.Info(ctx, fmt.Sprintf("Config state: %+v", state))
	
	// Debug to help see if this is being called with the correct job id
	if state.JobId.IsNull() {
		msg := "JobId is null, JobId must be specified"
		tflog.Error(ctx, msg)
		resp.Diagnostics.AddError("JobId is null", msg)
		return
	}

	jobIdValue := state.JobId.ValueInt64()
	tflog.Info(ctx, fmt.Sprintf("Reading job with ID: %d", jobIdValue))
	
	// Convert the Int64 to a string for the API call
	jobId := strconv.FormatInt(jobIdValue, 10)

	tflog.Info(ctx, fmt.Sprintf("Calling API for job with ID: %s", jobId))
	job, err := j.client.GetJob(jobId)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error getting job: %s", err.Error()))
		resp.Diagnostics.AddError("Error getting the job", err.Error())
		return
	}
	tflog.Info(ctx, "Successfully retrieved job from API")

	// Handle job completion trigger condition if present
	var jobCompletionTriggerCondition *JobCompletionTrigger
	if job.JobCompletionTrigger != nil {
		jobCompletionTriggerCondition = &JobCompletionTrigger{
			Condition: JobCompletionTriggerCondition{
				JobID:     types.Int64Value(int64(job.JobCompletionTrigger.Condition.JobID)),
				ProjectID: types.Int64Value(int64(job.JobCompletionTrigger.Condition.ProjectID)),
				Statuses: lo.Map(
					job.JobCompletionTrigger.Condition.Statuses,
					func(status int, _ int) types.String {
						return types.StringValue(
							utils.JobCompletionTriggerConditionsMappingCodeHuman[status].(string),
						)
					},
				),
			},
		}
	}

	// Populate state with job data
	state.Execution = &JobExecution{
		TimeoutSeconds: types.Int64Value(int64(job.Execution.TimeoutSeconds)),
	}
	state.GenerateDocs = types.BoolValue(job.GenerateDocs)
	state.RunGenerateSources = types.BoolValue(job.RunGenerateSources)
	state.ID = types.Int64PointerValue(helper.IntPointerToInt64Pointer(job.ID))
	state.ProjectID = types.Int64Value(int64(job.ProjectId))
	state.EnvironmentID = types.Int64Value(int64(job.EnvironmentId))
	state.Name = types.StringValue(job.Name)
	state.Description = types.StringValue(job.Description)
	state.DbtVersion = types.StringPointerValue(job.DbtVersion)
	state.ExecuteSteps = helper.SliceStringToSliceTypesString(job.ExecuteSteps)
	state.DeferringJobDefinitionID = types.Int64PointerValue(helper.IntPointerToInt64Pointer(job.DeferringJobId))
	state.DeferringEnvironmentID = types.Int64PointerValue(helper.IntPointerToInt64Pointer(job.DeferringEnvironmentId))
	state.Triggers = &JobTriggers{
		GithubWebhook:      types.BoolValue(job.Triggers.GithubWebhook),
		GitProviderWebhook: types.BoolValue(job.Triggers.GitProviderWebhook),
		Schedule:           types.BoolValue(job.Triggers.Schedule),
		OnMerge:            types.BoolValue(job.Triggers.OnMerge),
	}
	state.Settings = &JobSettings{
		Threads:    types.Int64Value(int64(job.Settings.Threads)),
		TargetName: types.StringValue(job.Settings.TargetName),
	}
	state.Schedule = &JobSchedule{
		Cron: types.StringValue(job.Schedule.Cron),
	}
	state.JobType = types.StringValue(job.JobType)
	state.TriggersOnDraftPr = types.BoolValue(job.TriggersOnDraftPR)
	state.RunCompareChanges = types.BoolValue(job.RunCompareChanges)
	state.JobCompletionTriggerCondition = jobCompletionTriggerCondition

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}