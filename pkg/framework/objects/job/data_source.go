package job

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
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
	var data SingleJobDataSourceModel

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

func (j *jobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state SingleJobDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	jobIdValue := state.JobId.ValueInt64()

	jobId := strconv.FormatInt(jobIdValue, 10)

	job, err := j.client.GetJob(jobId)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error getting job: %s", err.Error()))
		resp.Diagnostics.AddError("Error getting the job", err.Error())
		return
	}

	state.Execution = &JobExecution{
		TimeoutSeconds: types.Int64Value(int64(job.Execution.TimeoutSeconds)),
	}
	state.TimeoutSeconds = types.Int64Value(int64(job.Execution.TimeoutSeconds))
	state.GenerateDocs = types.BoolValue(job.GenerateDocs)
	state.RunGenerateSources = types.BoolValue(job.RunGenerateSources)
	state.ID = types.Int64PointerValue(helper.IntPointerToInt64Pointer(job.ID))
	state.ProjectID = types.Int64Value(int64(job.ProjectId))
	state.EnvironmentID = types.Int64Value(int64(job.EnvironmentId))
	state.Name = types.StringValue(job.Name)
	state.Description = types.StringValue(job.Description)
	state.DbtVersion = types.StringPointerValue(job.DbtVersion)
	state.ExecuteSteps = helper.SliceStringToSliceTypesString(job.ExecuteSteps)
	state.DeferringJobId = types.Int64PointerValue(helper.IntPointerToInt64Pointer(job.DeferringJobId))
	state.SelfDeferring = types.BoolValue(job.DeferringJobId != nil && *job.DeferringJobId == *job.ID)
	state.DeferringEnvironmentID = types.Int64PointerValue(helper.IntPointerToInt64Pointer(job.DeferringEnvironmentId))

	if job.ForceNodeSelection != nil {
		state.ForceNodeSelection = types.BoolValue(*job.ForceNodeSelection)
	} else {
		state.ForceNodeSelection = types.BoolNull()
	}

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
	state.RunCompareChanges = types.BoolPointerValue(job.RunCompareChanges)

	if job.JobCompletionTrigger != nil {
		state.JobCompletionTriggerCondition = []*JobCompletionTriggerCondition{
			{
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

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
