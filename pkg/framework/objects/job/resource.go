package job

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

func (j *jobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	jobID, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Job ID",
			fmt.Sprintf("Could not parse job_id from import ID %q: %v", req.ID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), jobID)...)
}

func (j *jobResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	j.client = req.ProviderData.(*dbt_cloud.Client)
}

func (j *jobResource) Create(ctx context.Context,req resource.CreateRequest,resp *resource.CreateResponse) {
	var plan JobResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId := plan.ProjectID
	environmentId := plan.EnvironmentID
	name := plan.Name.ValueString()
	description := plan.Description.ValueString()

	executeSteps := make([]string, len(plan.ExecuteSteps))
	for i, step := range plan.ExecuteSteps {
		executeSteps[i] = step.ValueString()
	}

	var dbtVersion *string
	if !plan.DbtVersion.IsNull() {
		dbtVersionValue := plan.DbtVersion.ValueString()
		dbtVersion = &dbtVersionValue
	}

	isActive := plan.IsActive.ValueBool()
	triggers := map[string]any{
		"github_webhook":       plan.Triggers.GithubWebhook.ValueBool(),
		"git_provider_webhook": plan.Triggers.GitProviderWebhook.ValueBool(),
		"schedule":           plan.Triggers.Schedule.ValueBool(),
		"on_merge":           plan.Triggers.OnMerge.ValueBool(),
	}
	numThreads := int(plan.NumThreads.ValueInt64())
	targetName := plan.TargetName.ValueString()
	generateDocs := plan.GenerateDocs.ValueBool()
	runGenerateSources := plan.RunGenerateSources.ValueBool()
	scheduleType := plan.ScheduleType.ValueString()
	scheduleInterval := int(plan.ScheduleInterval.ValueInt64())

	scheduleHours := make([]int, len(plan.ScheduleHours))
	for i, hour := range plan.ScheduleHours {
		scheduleHours[i] = int(hour.ValueInt64())
	}

	scheduleDays := make([]int, len(plan.ScheduleDays))
	for i, day := range plan.ScheduleDays {
		scheduleDays[i] = int(day.ValueInt64())
	}

	scheduleCron := plan.ScheduleCron.ValueString()

	var deferringJobId *int
	if !plan.DeferringJobId.IsNull() {
		deferringJobId = helper.Int64ToIntPointer(plan.DeferringJobId.ValueInt64())
	}

	var deferringEnvironmentID *int
	if !plan.DeferringEnvironmentID.IsNull() {
		deferringEnvironmentID = helper.Int64ToIntPointer(plan.DeferringEnvironmentID.ValueInt64())
	}

	selfDeferring := plan.SelfDeferring.ValueBool()
	timeoutSeconds := int(plan.TimeoutSeconds.ValueInt64())
	triggersOnDraftPR := plan.TriggersOnDraftPr.ValueBool()

	var jobCompletionTriggerCondition map[string]any
	if plan.JobCompletionTriggerCondition != nil {
		condition := plan.JobCompletionTriggerCondition.Condition
		statuses := make([]int, len(condition.Statuses))
		for i, status := range condition.Statuses {
			statuses[i] = utils.JobCompletionTriggerConditionsMappingHumanCode[status.ValueString()]
		}
		jobCompletionTriggerCondition = map[string]any{
			"job_id":     int(condition.JobID.ValueInt64()),
			"project_id": int(condition.ProjectID.ValueInt64()),
			"statuses":   statuses,
		}
	}

	runCompareChanges := plan.RunCompareChanges.ValueBool()
	runLint := plan.RunLint.ValueBool()
	errorsOnLintFailure := plan.ErrorsOnLintFailure.ValueBool()
	jobType := plan.JobType.ValueString()
	compareChangesFlags := plan.CompareChangesFlags.ValueString()

	createDbtVersion := ""
	if dbtVersion != nil {
		createDbtVersion = *dbtVersion
	}
	createDeferringJobID := 0
	if deferringJobId != nil {
		createDeferringJobID = *deferringJobId
	}
	createDeferringEnvironmentID := 0
	if deferringEnvironmentID != nil {
		createDeferringEnvironmentID = *deferringEnvironmentID
	}

	createdJob, err := j.client.CreateJob(int(projectId.ValueInt64()),
		int(environmentId.ValueInt64()),
		name,
		description,
		executeSteps,
		createDbtVersion,
		isActive,
		triggers,
		numThreads,
		targetName,
		generateDocs,
		runGenerateSources,
		scheduleType,
		scheduleInterval,
		scheduleHours,
		scheduleDays,
		scheduleCron,
		createDeferringJobID,
		createDeferringEnvironmentID,
		selfDeferring,
		timeoutSeconds,
		triggersOnDraftPR,
		jobCompletionTriggerCondition,
		runCompareChanges,
		runLint,
		errorsOnLintFailure,
		jobType,
		compareChangesFlags,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating job",
			"Could not create job, unexpected error: "+err.Error(),
		)
		return
	}

	if createdJob != nil && createdJob.ID != nil {
		plan.ID = types.Int64Value(int64(*createdJob.ID))
		plan.JobId= types.Int64Value(int64(*createdJob.ID))
	} else {
		resp.Diagnostics.AddError(
			"Error creating job",
			"Created job or its ID is unexpectedly nil",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (j *jobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state JobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := state.ID.ValueInt64()
	jobIDStr := strconv.FormatInt(jobID, 10)

	job, err := j.client.GetJob(jobIDStr)
	if err != nil {

		if strings.HasPrefix(err.Error(), "resource-not-found") {
			return
		}
		resp.Diagnostics.AddError("Client Error", "Unable to retrieve job before deletion")
		return
	}

	job.State = dbt_cloud.STATE_DELETED
	_, err = j.client.UpdateJob(jobIDStr, *job)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to delete job")
		return
	}
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

func (j *jobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state JobResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	jobID := state.ID.ValueInt64()
	jobIDStr := strconv.FormatInt(jobID, 10)

	job, err := j.client.GetJob(jobIDStr)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving job",
			"Could not retrieve job with ID "+jobIDStr+": "+err.Error(),
		)
		return
	}

	job.ProjectId = int(plan.ProjectID.ValueInt64())
	job.EnvironmentId = int(plan.EnvironmentID.ValueInt64())
	job.Name = plan.Name.ValueString()
	job.Description = plan.Description.ValueString()

	if plan.DbtVersion.IsNull() {
		job.DbtVersion = nil
	} else {
		dbtVersion := plan.DbtVersion.ValueString()
		job.DbtVersion = &dbtVersion
	}

	job.Settings.Threads = int(plan.NumThreads.ValueInt64())
	job.Settings.TargetName = plan.TargetName.ValueString()
	job.GenerateDocs = plan.GenerateDocs.ValueBool()
	job.RunGenerateSources = plan.RunGenerateSources.ValueBool()

	executeSteps := make([]string, len(plan.ExecuteSteps))
	for i, step := range plan.ExecuteSteps {
		executeSteps[i] = step.ValueString()
	}
	job.ExecuteSteps = executeSteps

	if plan.Triggers != nil {
		job.Triggers.GithubWebhook = plan.Triggers.GithubWebhook.ValueBool()
		job.Triggers.GitProviderWebhook = plan.Triggers.GitProviderWebhook.ValueBool()
		job.Triggers.Schedule = plan.Triggers.Schedule.ValueBool()
		job.Triggers.OnMerge = plan.Triggers.OnMerge.ValueBool()
	}

	scheduleType := plan.ScheduleType.ValueString()
	job.Schedule.Date.Type = scheduleType

	scheduleInterval := int(plan.ScheduleInterval.ValueInt64())
	job.Schedule.Time.Interval = scheduleInterval

	if len(plan.ScheduleHours) > 0 {
		scheduleHours := make([]int, len(plan.ScheduleHours))
		for i, hour := range plan.ScheduleHours {
			scheduleHours[i] = int(hour.ValueInt64())
		}
		job.Schedule.Time.Hours = &scheduleHours
		job.Schedule.Time.Type = "at_exact_hours" // Explicitly set type when hours are provided
		job.Schedule.Time.Interval = 0 // Reset interval when using specific hours
	} else {
		job.Schedule.Time.Hours = nil
		job.Schedule.Time.Type = "every_hour" // Reset to default type if hours are removed
		job.Schedule.Time.Interval = scheduleInterval // Ensure interval is set if hours are nil
	}

	if len(plan.ScheduleDays) > 0 {
		scheduleDays := make([]int, len(plan.ScheduleDays))
		for i, day := range plan.ScheduleDays {
			scheduleDays[i] = int(day.ValueInt64())
		}
		job.Schedule.Date.Days = &scheduleDays
	} else {
		job.Schedule.Date.Days = nil
	}

	if plan.ScheduleCron.IsNull() || plan.ScheduleCron.ValueString() == "" {
		job.Schedule.Date.Cron = nil
	} else {
		scheduleCron := plan.ScheduleCron.ValueString()
		job.Schedule.Date.Cron = &scheduleCron
	}

	if scheduleType == "days_of_week" || scheduleType == "every_day" {
		job.Schedule.Date.Cron = nil
	}
	if scheduleType == "custom_cron" || scheduleType == "every_day" {
		job.Schedule.Date.Days = nil
	}


	selfDeferring := plan.SelfDeferring.ValueBool()
	if selfDeferring {
		deferringJobID := int(jobID)
		job.DeferringJobId = &deferringJobID
		job.DeferringEnvironmentId = nil // Self deferring is mutually exclusive with environment deferring
	} else {
		if plan.DeferringJobId.IsNull() || plan.DeferringJobId.ValueInt64() == 0 {
			job.DeferringJobId = nil
		} else {
			deferringJobId := int(plan.DeferringJobId.ValueInt64())
			job.DeferringJobId = &deferringJobId
		}

		if plan.DeferringEnvironmentID.IsNull() || plan.DeferringEnvironmentID.ValueInt64() == 0 {
			job.DeferringEnvironmentId = nil
		} else {
			deferringEnvId := int(plan.DeferringEnvironmentID.ValueInt64())
			job.DeferringEnvironmentId = &deferringEnvId
		}
	}

	job.Execution.TimeoutSeconds = int(plan.TimeoutSeconds.ValueInt64())
	job.TriggersOnDraftPR = plan.TriggersOnDraftPr.ValueBool()
	job.RunCompareChanges = plan.RunCompareChanges.ValueBool()
	job.RunLint = plan.RunLint.ValueBool()
	job.ErrorsOnLintFailure = plan.ErrorsOnLintFailure.ValueBool()
	job.CompareChangesFlags = plan.CompareChangesFlags.ValueString()

	if plan.JobCompletionTriggerCondition == nil {
		job.JobCompletionTrigger = nil
	} else {
		condition := plan.JobCompletionTriggerCondition.Condition
		statuses := make([]int, len(condition.Statuses))
		for i, status := range condition.Statuses {
			statuses[i] = utils.JobCompletionTriggerConditionsMappingHumanCode[status.ValueString()]
		}
		jobCondTrigger := dbt_cloud.JobCompletionTrigger{
			Condition: dbt_cloud.JobCompletionTriggerCondition{
				JobID:     int(condition.JobID.ValueInt64()),
				ProjectID: int(condition.ProjectID.ValueInt64()),
				Statuses:  statuses,
			},
		}
		job.JobCompletionTrigger = &jobCondTrigger
	}

	_, err = j.client.UpdateJob(jobIDStr, *job)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating job",
			"Could not update job with ID "+jobIDStr+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
