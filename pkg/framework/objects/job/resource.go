package job

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &jobResource{}
	_ resource.ResourceWithConfigure   = &jobResource{}
	_ resource.ResourceWithImportState = &jobResource{}
)

type jobResource struct {
	client *dbt_cloud.Client
}

func JobResource() resource.Resource {
	return &jobResource{}
}

func (j *jobResource) ImportState(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse) {
	panic("unimplemented")
}

func (j *jobResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	j.client = req.ProviderData.(*dbt_cloud.Client)
}

func (j *jobResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
	panic("unimplemented")
}

func (j *jobResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	panic("unimplemented")
}

func (j *jobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

func (j *jobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state JobResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	jobID := state.ID.ValueInt64()
	jobIDStr := strconv.FormatInt(jobID, 10)

	retrievedJob, err := j.client.GetJob(jobIDStr)

	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The job was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the job", err.Error())
		return
	}

	state.ID = types.Int64Value(int64(*retrievedJob.ID))
	state.ProjectID = types.Int64Value(int64(retrievedJob.ProjectId))
	state.EnvironmentID = types.Int64Value(int64(retrievedJob.EnvironmentId))
	state.Name = types.StringValue(retrievedJob.Name)
	state.Description = types.StringValue(retrievedJob.Description)
	state.ExecuteSteps = helper.SliceStringToSliceTypesString(retrievedJob.ExecuteSteps)
	state.DbtVersion = types.StringValue(*retrievedJob.DbtVersion)
	state.IsActive = types.BoolValue(retrievedJob.State == 1)
	state.NumThreads = types.Int64Value(int64(retrievedJob.Settings.Threads))
	state.TargetName = types.StringValue(retrievedJob.Settings.TargetName)
	state.GenerateDocs = types.BoolValue(retrievedJob.GenerateDocs)
	state.RunGenerateSources = types.BoolValue(retrievedJob.RunGenerateSources)
	state.ScheduleType = types.StringValue(retrievedJob.Schedule.Date.Type)

	schedule := 1
	if retrievedJob.Schedule.Time.Interval > 0 {
		schedule = retrievedJob.Schedule.Time.Interval
	}
	state.ScheduleInterval = types.Int64Value(int64(schedule))

	if retrievedJob.Schedule.Time.Hours != nil {
		state.ScheduleHours = helper.SliceIntToSliceTypesInt64(*retrievedJob.Schedule.Time.Hours)
	} else {
		state.ScheduleHours = []types.Int64{}
	}
	
	if retrievedJob.Schedule.Date.Days != nil {
		state.ScheduleDays = helper.SliceIntToSliceTypesInt64(*retrievedJob.Schedule.Date.Days)
	} else {
		state.ScheduleDays = []types.Int64{}
	}
	state.ScheduleCron = types.StringValue(*retrievedJob.Schedule.Date.Cron)

	selfDeferring := retrievedJob.DeferringJobId != nil && strconv.Itoa(*retrievedJob.DeferringJobId) == jobIDStr
	state.SelfDeferring = types.BoolValue(selfDeferring)

	if !selfDeferring {
		state.DeferringJobId = types.Int64Value(int64(*retrievedJob.DeferringJobId))
	}

	state.DeferringEnvironmentID = types.Int64Value(int64(*retrievedJob.DeferringEnvironmentId))
	state.TimeoutSeconds = types.Int64Value(int64(retrievedJob.Execution.TimeoutSeconds))



	// for now, we allow people to keep the triggers.custom_branch_only config even if the parameter was deprecated in the API
	// we set the state to the current config value, so it doesn't do anything
	// todo check the custom branch stuff and on merge

	// var triggers map[string]interface{}
	// triggersInput, _ := json.Marshal(job.Triggers)
	// json.Unmarshal(triggersInput, &triggers)

	// // for now, we allow people to keep the triggers.custom_branch_only config even if the parameter was deprecated in the API
	// // we set the state to the current config value, so it doesn't do anything
	// listedTriggers := d.Get("triggers").(map[string]interface{})
	// listedCustomBranchOnly, ok := listedTriggers["custom_branch_only"].(bool)
	// if ok {
	// 	triggers["custom_branch_only"] = listedCustomBranchOnly
	// }

	// // we remove triggers.on_merge if it is not set in the config and it is set to false in the remote
	// // that way it works if people don't define it, but also works to import jobs that have it set to true
	// // TODO: remove this when we make on_merge mandatory
	// _, ok = listedTriggers["on_merge"].(bool)
	// noOnMergeConfig := !ok
	// onMergeRemoteVal, _ := triggers["on_merge"].(bool)
	// onMergeRemoteFalse := !onMergeRemoteVal
	// if noOnMergeConfig && onMergeRemoteFalse {
	// 	delete(triggers, "on_merge")
	// }



	state.Triggers = &JobTriggers{
		GithubWebhook:      types.BoolValue(retrievedJob.Triggers.GithubWebhook),
		GitProviderWebhook: types.BoolValue(retrievedJob.Triggers.GitProviderWebhook),
		Schedule:           types.BoolValue(retrievedJob.Triggers.Schedule),
		OnMerge:            types.BoolValue(retrievedJob.Triggers.OnMerge),
	}





	
	state.RunCompareChanges = types.BoolValue(retrievedJob.RunCompareChanges)
	state.CompareChangesFlags = types.StringValue(retrievedJob.CompareChangesFlags)
	state.RunLint = types.BoolValue(retrievedJob.RunLint)
	state.ErrorsOnLintFailure = types.BoolValue(retrievedJob.ErrorsOnLintFailure)
	state.JobType = types.StringValue(retrievedJob.JobType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (j *jobResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	panic("unimplemented")
}
