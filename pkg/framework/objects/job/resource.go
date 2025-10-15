package job

import (
	"context"
	"encoding/json"
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
	_ resource.ResourceWithModifyPlan  = &jobResource{}
)

type jobResource struct {
	client *dbt_cloud.Client
}

func (j *jobResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Don't do anything on resource creation or deletion
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state JobResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Skip checks if necessary fields are null
	if plan.Triggers == nil || state.Triggers == nil {
		return
	}

	// if we change the job type (CI, merge or "empty"), we need to recreate the job as dbt Cloud doesn't allow updating them
	// the job type is determined by the triggers
	if plan.Triggers != nil && state.Triggers != nil {
		oldCI := state.Triggers.GithubWebhook.ValueBool() || state.Triggers.GitProviderWebhook.ValueBool()
		oldOnMerge := state.Triggers.OnMerge.ValueBool()

		oldType := ""
		if oldCI {
			oldType = "ci"
		} else if oldOnMerge {
			oldType = "merge"
		}

		newCI := plan.Triggers.GithubWebhook.ValueBool() || plan.Triggers.GitProviderWebhook.ValueBool()
		newOnMerge := plan.Triggers.OnMerge.ValueBool()

		newType := ""
		if newCI {
			newType = "ci"
		} else if newOnMerge {
			newType = "merge"
		}

		if oldType != newType {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("triggers"))
		}
	}
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
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_id"), jobID)...)
}

func (j *jobResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	j.client = req.ProviderData.(*dbt_cloud.Client)
}

func (j *jobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
		"schedule":             plan.Triggers.Schedule.ValueBool(),
		"on_merge":             plan.Triggers.OnMerge.ValueBool(),
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
	runCompareChanges := plan.RunCompareChanges.ValueBool()
	runLint := plan.RunLint.ValueBool()
	errorsOnLintFailure := plan.ErrorsOnLintFailure.ValueBool()

	var jobType string
	if !plan.JobType.IsNull() {
		jobType = plan.JobType.ValueString()
	}

	compareChangesFlags := plan.CompareChangesFlags.ValueString()

	var jobCompletionTriggerCondition map[string]any

	if len(plan.JobCompletionTriggerCondition) != 0 {
		condition := plan.JobCompletionTriggerCondition[0]
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

	// Defensive check: ensure createdJob and its ID are not nil before dereferencing
	if createdJob == nil {
		resp.Diagnostics.AddError(
			"Error creating job",
			"Job creation returned nil response without an error. This may indicate a permissions issue or an API problem.",
		)
		return
	}

	if createdJob.ID == nil {
		resp.Diagnostics.AddError(
			"Error creating job",
			"Job creation returned a response without a job ID. This may indicate a permissions issue or an API problem.",
		)
		return
	}

	plan.ID = types.Int64Value(int64(*createdJob.ID))
	plan.JobId = types.Int64Value(int64(*createdJob.ID))

	if createdJob.JobType != "" {
		plan.JobType = types.StringValue(createdJob.JobType)
	} else {
		plan.JobType = types.StringNull()
	}

	jobIDStr := strconv.FormatInt(int64(*createdJob.ID), 10)

	// Check if DeferringJobId is set and matches this job's ID for self-deferring
	createdSelfDeferring := false
	if createdJob.DeferringJobId != nil {
		createdSelfDeferring = strconv.Itoa(*createdJob.DeferringJobId) == jobIDStr
	}
	plan.SelfDeferring = types.BoolValue(createdSelfDeferring)

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (j *jobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state JobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure ID is not null before accessing
	if state.ID.IsNull() {
		resp.Diagnostics.AddError("Client Error", "Job ID is null")
		return
	}

	jobID := state.ID.ValueInt64()
	jobIDStr := strconv.FormatInt(jobID, 10)

	job, err := j.client.GetJob(jobIDStr)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			return
		}
		resp.Diagnostics.AddError("Client Error", "Unable to retrieve job before deletion: "+err.Error())
		return
	}

	job.State = dbt_cloud.STATE_DELETED
	_, err = j.client.UpdateJob(jobIDStr, *job)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to delete job: "+err.Error())
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
	state.JobId = types.Int64Value(int64(*retrievedJob.ID))
	state.ProjectID = types.Int64Value(int64(retrievedJob.ProjectId))
	state.EnvironmentID = types.Int64Value(int64(retrievedJob.EnvironmentId))
	state.Name = types.StringValue(retrievedJob.Name)
	state.Description = types.StringValue(retrievedJob.Description)
	state.ExecuteSteps = helper.SliceStringToSliceTypesString(retrievedJob.ExecuteSteps)

	if retrievedJob.DbtVersion != nil {
		state.DbtVersion = types.StringValue(*retrievedJob.DbtVersion)
	} else {
		state.DbtVersion = types.StringNull()
	}

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
		var scheduleHoursNull []types.Int64
		state.ScheduleHours = scheduleHoursNull
	}

	if retrievedJob.Schedule.Date.Days != nil {
		state.ScheduleDays = helper.SliceIntToSliceTypesInt64(*retrievedJob.Schedule.Date.Days)
	} else {
		var scheduleDaysNull []types.Int64
		state.ScheduleDays = scheduleDaysNull
	}

	if retrievedJob.Schedule.Date.Cron != nil &&
		retrievedJob.Schedule.Date.Type != "interval_cron" { // for interval_cron, the cron expression is auto generated in the code
		state.ScheduleCron = types.StringValue(*retrievedJob.Schedule.Date.Cron)
	} else {
		state.ScheduleCron = types.StringNull()
	}

	selfDeferring := retrievedJob.DeferringJobId != nil && strconv.Itoa(*retrievedJob.DeferringJobId) == jobIDStr

	if retrievedJob.DeferringJobId != nil && !selfDeferring {
		state.DeferringJobId = types.Int64Value(int64(*retrievedJob.DeferringJobId))
	} else {
		state.DeferringJobId = types.Int64Null()
	}

	if retrievedJob.DeferringEnvironmentId != nil {
		state.DeferringEnvironmentID = types.Int64Value(int64(*retrievedJob.DeferringEnvironmentId))
	} else {
		state.DeferringEnvironmentID = types.Int64Null()
	}

	state.SelfDeferring = types.BoolValue(selfDeferring)
	state.TimeoutSeconds = types.Int64Value(int64(retrievedJob.Execution.TimeoutSeconds))

	var triggers map[string]interface{}
	triggersInput, _ := json.Marshal(retrievedJob.Triggers)
	json.Unmarshal(triggersInput, &triggers)

	// for now, we allow people to keep the triggers.custom_branch_only config even if the parameter was deprecated in the API
	// we set the state to the current config value, so it doesn't do anything
	var customBranchValue types.Bool
	diags := req.State.GetAttribute(ctx, path.Root("triggers").AtMapKey("custom_branch_only"), &customBranchValue)

	if !diags.HasError() && !customBranchValue.IsNull() {
		triggers["custom_branch_only"] = customBranchValue.ValueBool()
	}

	// we remove triggers.on_merge if it is not set in the config and it is set to false in the remote
	// that way it works if people don't define it, but also works to import jobs that have it set to true
	// TODO: remove this when we make on_merge mandatory
	var onMergeValue types.Bool
	hasOnMergeAttr := !req.State.GetAttribute(ctx, path.Root("triggers").AtMapKey("on_merge"), &onMergeValue).HasError()
	noOnMergeConfig := !hasOnMergeAttr || onMergeValue.IsNull()

	onMergeRemoteVal, _ := triggers["on_merge"].(bool)
	onMergeRemoteFalse := !onMergeRemoteVal

	if noOnMergeConfig && onMergeRemoteFalse {
		delete(triggers, "on_merge")
	}

	state.Triggers = &JobTriggers{
		GithubWebhook:      types.BoolValue(retrievedJob.Triggers.GithubWebhook),
		GitProviderWebhook: types.BoolValue(retrievedJob.Triggers.GitProviderWebhook),
		Schedule:           types.BoolValue(retrievedJob.Triggers.Schedule),
		OnMerge:            types.BoolValue(retrievedJob.Triggers.OnMerge),
	}

	state.TriggersOnDraftPr = types.BoolValue(retrievedJob.TriggersOnDraftPR)
	if retrievedJob.JobCompletionTrigger != nil {
		statusesStr := make([]types.String, 0)
		for _, status := range retrievedJob.JobCompletionTrigger.Condition.Statuses {
			statusStr := utils.JobCompletionTriggerConditionsMappingCodeHuman[status].(string)
			statusesStr = append(statusesStr, types.StringValue(statusStr))
		}

		state.JobCompletionTriggerCondition = []*JobCompletionTriggerCondition{
			{
				JobID:     types.Int64Value(int64(retrievedJob.JobCompletionTrigger.Condition.JobID)),
				ProjectID: types.Int64Value(int64(retrievedJob.JobCompletionTrigger.Condition.ProjectID)),
				Statuses:  statusesStr,
			},
		}
	} else {
		state.JobCompletionTriggerCondition = nil
	}

	state.RunCompareChanges = types.BoolValue(retrievedJob.RunCompareChanges)
	state.CompareChangesFlags = types.StringValue(retrievedJob.CompareChangesFlags)
	state.RunLint = types.BoolValue(retrievedJob.RunLint)
	state.ErrorsOnLintFailure = types.BoolValue(retrievedJob.ErrorsOnLintFailure)

	if retrievedJob.JobType != "" {
		state.JobType = types.StringValue(retrievedJob.JobType)
	} else {
		state.JobType = types.StringNull()
	}

	if state.SelfDeferring.IsNull() {
		state.SelfDeferring = types.BoolValue(selfDeferring)
	}

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
	job.RunGenerateSources = plan.RunGenerateSources.ValueBool()
	job.GenerateDocs = plan.GenerateDocs.ValueBool()

	executeSteps := make([]string, len(plan.ExecuteSteps))
	for i, step := range plan.ExecuteSteps {
		executeSteps[i] = step.ValueString()
	}
	job.ExecuteSteps = executeSteps

	// todo check if trigger handling is ok
	if plan.Triggers != nil {
		job.Triggers.GithubWebhook = plan.Triggers.GithubWebhook.ValueBool()
		job.Triggers.GitProviderWebhook = plan.Triggers.GitProviderWebhook.ValueBool()
		job.Triggers.Schedule = plan.Triggers.Schedule.ValueBool()
		job.Triggers.OnMerge = plan.Triggers.OnMerge.ValueBool()
	}

	scheduleInterval := int(plan.ScheduleInterval.ValueInt64())
	job.Schedule.Time.Interval = scheduleInterval

	if len(plan.ScheduleHours) > 0 {
		scheduleHours := make([]int, len(plan.ScheduleHours))
		for i, hour := range plan.ScheduleHours {
			scheduleHours[i] = int(hour.ValueInt64())
		}
		job.Schedule.Time.Hours = &scheduleHours
		job.Schedule.Time.Type = "at_exact_hours"
		job.Schedule.Time.Interval = 0
	} else {
		job.Schedule.Time.Hours = nil
		job.Schedule.Time.Type = "every_hour"
		job.Schedule.Time.Interval = scheduleInterval
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

	// we set this after the subfields to remove the fields not matching the schedule type
	// if it was before, some of those fields would be set again
	scheduleType := plan.ScheduleType.ValueString()
	job.Schedule.Date.Type = scheduleType

	if scheduleType == "days_of_week" || scheduleType == "every_day" {
		job.Schedule.Date.Cron = nil
	}
	if scheduleType == "custom_cron" || scheduleType == "every_day" {
		job.Schedule.Date.Days = nil
	}

	if plan.DeferringEnvironmentID.IsNull() || plan.DeferringEnvironmentID.ValueInt64() == 0 {
		job.DeferringEnvironmentId = nil
	} else {
		deferringEnvId := int(plan.DeferringEnvironmentID.ValueInt64())
		job.DeferringEnvironmentId = &deferringEnvId
	}

	// If self_deferring has been toggled to true, set deferring_job_id as own ID
	// Otherwise, set it back to what deferring_job_id specifies it to be
	selfDeferring := plan.SelfDeferring.ValueBool()
	if selfDeferring {
		deferringJobID := int(jobID)
		job.DeferringJobId = &deferringJobID
	} else {
		if plan.DeferringJobId.IsNull() || plan.DeferringJobId.ValueInt64() == 0 {
			job.DeferringJobId = nil
		} else {
			deferringJobId := int(plan.DeferringJobId.ValueInt64())
			job.DeferringJobId = &deferringJobId
		}
	}

	job.Execution.TimeoutSeconds = int(plan.TimeoutSeconds.ValueInt64())
	job.TriggersOnDraftPR = plan.TriggersOnDraftPr.ValueBool()

	if len(plan.JobCompletionTriggerCondition) == 0 {
		job.JobCompletionTrigger = nil
	} else {
		condition := plan.JobCompletionTriggerCondition[0]
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

	job.RunCompareChanges = plan.RunCompareChanges.ValueBool()
	job.RunLint = plan.RunLint.ValueBool()
	job.ErrorsOnLintFailure = plan.ErrorsOnLintFailure.ValueBool()
	job.CompareChangesFlags = plan.CompareChangesFlags.ValueString()

	// Capture what's changing for better error messages
	oldEnvID := state.EnvironmentID.ValueInt64()
	newEnvID := plan.EnvironmentID.ValueInt64()
	oldName := state.Name.ValueString()
	newName := plan.Name.ValueString()

	updatedJob, err := j.client.UpdateJob(jobIDStr, *job)
	if err != nil {
		// Build a well-formatted, context-aware error message
		var errorMsg strings.Builder
		errorMsg.WriteString(fmt.Sprintf("Could not update job with ID %s\n\n", jobIDStr))
		errorMsg.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))

		// Add context about what was being changed
		var changes []string
		if oldEnvID != newEnvID {
			changes = append(changes, fmt.Sprintf("  • environment_id: %d → %d", oldEnvID, newEnvID))
		}
		if oldName != newName {
			changes = append(changes, fmt.Sprintf("  • name: '%s' → '%s'", oldName, newName))
		}

		if len(changes) > 0 {
			errorMsg.WriteString("\nAttempted changes:\n")
			errorMsg.WriteString(strings.Join(changes, "\n"))

			// If environment is changing and it's a permission error, add extra context
			if oldEnvID != newEnvID && (strings.Contains(err.Error(), "permission") || strings.Contains(err.Error(), "forbidden") || strings.Contains(err.Error(), "resource-not-found-permissions")) {
				errorMsg.WriteString(fmt.Sprintf("\n\nℹ️  Note: The API token may not have write access to environment %d.\nEnvironment-level permissions are required to move jobs between environments.", newEnvID))
			}
		}

		resp.Diagnostics.AddError(
			"Error updating job",
			errorMsg.String(),
		)
		return
	}

	if updatedJob.JobType != "" {
		plan.JobType = types.StringValue(updatedJob.JobType)
	} else {
		plan.JobType = types.StringNull()
	}

	updatedJobIDStr := strconv.FormatInt(jobID, 10)
	updatedSelfDeferring := updatedJob.DeferringJobId != nil && strconv.Itoa(*updatedJob.DeferringJobId) == updatedJobIDStr
	plan.SelfDeferring = types.BoolValue(updatedSelfDeferring)

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
