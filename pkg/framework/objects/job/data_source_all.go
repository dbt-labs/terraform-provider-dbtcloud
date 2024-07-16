package job

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

var (
	_ datasource.DataSource                   = &jobsDataSource{}
	_ datasource.DataSourceWithConfigure      = &jobsDataSource{}
	_ datasource.DataSourceWithValidateConfig = &jobsDataSource{}
)

func JobsDataSource() datasource.DataSource {
	return &jobsDataSource{}
}

type jobsDataSource struct {
	client *dbt_cloud.Client
}

func (d *jobsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_jobs"
}

func (d *jobsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var config JobsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	var projectID int
	if config.ProjectID.IsNull() {
		projectID = 0
	} else {
		projectID = int(config.ProjectID.ValueInt64())
	}
	var environmentID int
	if config.EnvironmentID.IsNull() {
		environmentID = 0
	} else {
		environmentID = int(config.EnvironmentID.ValueInt64())
	}

	apiJobs, err := d.client.GetAllJobs(projectID, environmentID)

	if err != nil {
		resp.Diagnostics.AddError(
			"Issue when retrieving jobs",
			err.Error(),
		)
		return
	}

	state := config

	allJobs := []JobDataSourceModel{}
	for _, job := range apiJobs {

		// we need to handle the case the condition is nil
		var jobCompletionTriggerCondition *JobCompletionTrigger
		if job.JobCompletionTrigger != nil {
			jobCompletionTriggerCondition = &JobCompletionTrigger{
				Condition: JobCompletionTriggerCondition{
					JobID: types.Int64Value(
						int64(job.JobCompletionTrigger.Condition.JobID),
					),
					ProjectID: types.Int64Value(
						int64(job.JobCompletionTrigger.Condition.ProjectID),
					),
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

		currentJob := JobDataSourceModel{
			Execution: JobExecution{
				TimeoutSeconds: types.Int64Value(int64(job.Execution.Timeout_Seconds)),
			},
			GenerateDocs:       types.BoolValue(job.Generate_Docs),
			RunGenerateSources: types.BoolValue(job.Run_Generate_Sources),
			ID: types.Int64PointerValue(
				helper.IntPointerToInt64Pointer(job.ID),
			),
			ProjectID:     types.Int64Value(int64(job.Project_Id)),
			EnvironmentID: types.Int64Value(int64(job.Environment_Id)),
			Name:          types.StringValue(job.Name),
			Description:   types.StringValue(job.Description),
			DbtVersion: types.StringPointerValue(
				job.Dbt_Version,
			),
			ExecuteSteps: helper.SliceStringToSliceTypesString(job.Execute_Steps),
			DeferringJobDefinitionID: types.Int64PointerValue(helper.IntPointerToInt64Pointer(
				job.Deferring_Job_Id),
			),
			DeferringEnvironmentID: types.Int64PointerValue(helper.IntPointerToInt64Pointer(
				job.DeferringEnvironmentId),
			),
			Triggers: JobTriggers{
				GithubWebhook:      types.BoolValue(job.Triggers.Github_Webhook),
				GitProviderWebhook: types.BoolValue(job.Triggers.GitProviderWebhook),
				Schedule:           types.BoolValue(job.Triggers.Schedule),
				OnMerge:            types.BoolValue(job.Triggers.OnMerge),
			},
			Settings: JobSettings{
				Threads:    types.Int64Value(int64(job.Settings.Threads)),
				TargetName: types.StringValue(job.Settings.Target_Name),
			},
			Schedule: JobSchedule{
				Cron: types.StringValue(job.Schedule.Cron),
			},
			JobType:           types.StringValue(job.JobType),
			TriggersOnDraftPr: types.BoolValue(job.TriggersOnDraftPR),
			Environment: JobEnvironment{
				ProjectID:      types.Int64Value(int64(job.Environment.Project_Id)),
				ID:             types.Int64Value(int64(*job.Environment.ID)),
				Name:           types.StringValue(job.Environment.Name),
				DeploymentType: types.StringPointerValue(job.Environment.DeploymentType),
				Type:           types.StringValue(job.Environment.Type),
			},
			JobCompletionTriggerCondition: jobCompletionTriggerCondition,
		}

		allJobs = append(allJobs, currentJob)
	}
	state.Jobs = allJobs

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *jobsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}
