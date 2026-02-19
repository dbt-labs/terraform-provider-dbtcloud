package job

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

// Job type constants matching the server-side JobType enum
const (
	JobTypeCI        = "ci"
	JobTypeMerge     = "merge"
	JobTypeScheduled = "scheduled"
	JobTypeOther     = "other"
	JobTypeAdaptive  = "adaptive"
)

type jobResource struct {
	client *dbt_cloud.Client
}

func debugJobTypeLog(runID, hypothesisID, location, message string, data map[string]any) {
	payload := map[string]any{
		"sessionId":    "cb43da",
		"runId":        runID,
		"hypothesisId": hypothesisID,
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	}
	logLine, err := json.Marshal(payload)
	if err != nil {
		return
	}
	f, err := os.OpenFile("/Users/operator/Documents/git/dbt-labs/terraform-dbtcloud-yaml/.cursor/debug-cb43da.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.Write(append(logLine, '\n'))
	_ = f.Close()
}

// validateJobTypeChange validates if a job type transition is allowed.
// This mirrors the server-side validation in _validate_job_type_change.
func validateJobTypeChange(prevJobType, newJobType string) error {
	// If no change, always allowed
	if prevJobType == newJobType {
		return nil
	}

	// If previous type is empty (not set), any new type is allowed
	if prevJobType == "" {
		return nil
	}

	// CI jobs can only stay CI
	if prevJobType == JobTypeCI && newJobType != JobTypeCI {
		return fmt.Errorf("the job type field for this job can only be set to 'ci'")
	}

	// Adaptive jobs can only stay adaptive
	if prevJobType == JobTypeAdaptive && newJobType != JobTypeAdaptive {
		return fmt.Errorf("the job type field for this job can only be set to 'adaptive'")
	}

	// Scheduled jobs can only change to scheduled or other
	if prevJobType == JobTypeScheduled && (newJobType == JobTypeCI || newJobType == JobTypeAdaptive) {
		return fmt.Errorf("the job type field for this job can only be set to 'scheduled' or 'other'")
	}

	// Other jobs can only change to scheduled or other
	if prevJobType == JobTypeOther && (newJobType == JobTypeCI || newJobType == JobTypeAdaptive) {
		return fmt.Errorf("the job type field for this job can only be set to 'scheduled' or 'other'")
	}

	// Merge jobs - treating similar to CI (cannot change away from merge)
	if prevJobType == JobTypeMerge && newJobType != JobTypeMerge {
		return fmt.Errorf("the job type field for this job can only be set to 'merge'")
	}

	return nil
}

func (j *jobResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.Plan.Raw.IsNull() {
		var plan JobResourceModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

		if plan.ValidateExecuteSteps.ValueBool() {
			executeSteps := make([]string, len(plan.ExecuteSteps))
			for i, step := range plan.ExecuteSteps {
				executeSteps[i] = step.ValueString()
			}

			if err := j.validateExecuteSteps(executeSteps); err != nil {
				resp.Diagnostics.AddError("Error validating execute steps", err.Error())
				return
			}
		}
	}

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

	// Validate job_type field changes if the field is being explicitly set
	// Note: If plan.JobType is set but state.JobType is null (first time setting it),
	// the validation will happen in Update against the actual server value
	// Skip validation if either value is empty (empty means "not explicitly set")
	if !plan.JobType.IsNull() && !state.JobType.IsNull() {
		prevJobType := state.JobType.ValueString()
		newJobType := plan.JobType.ValueString()

		// Only validate if both values are non-empty (explicitly set)
		if prevJobType != "" && newJobType != "" {
			if err := validateJobTypeChange(prevJobType, newJobType); err != nil {
				resp.Diagnostics.AddError(
					"Invalid job_type change",
					fmt.Sprintf("Cannot change job_type from '%s' to '%s': %s", prevJobType, newJobType, err.Error()),
				)
				return
			}

			// Force replacement only for CI and Adaptive type changes (API restrictions)
			// CI jobs cannot change to any other type
			// Adaptive jobs cannot change to any other type
			// Note: Merge jobs CAN have their on_merge trigger disabled without changing type
			// Scheduled/Other jobs can change between each other and to merge
			ciToNonCI := prevJobType == JobTypeCI && newJobType != JobTypeCI
			nonCIToCI := prevJobType != JobTypeCI && newJobType == JobTypeCI
			adaptiveToNonAdaptive := prevJobType == JobTypeAdaptive && newJobType != JobTypeAdaptive
			nonAdaptiveToAdaptive := prevJobType != JobTypeAdaptive && newJobType == JobTypeAdaptive

			if ciToNonCI || nonCIToCI || adaptiveToNonAdaptive || nonAdaptiveToAdaptive {
				resp.RequiresReplace = append(resp.RequiresReplace, path.Root("job_type"))
			}
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

	isActive := true // when being created, the job should be active by default
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

	// ForceNodeSelection: Only set if explicitly provided in config (not null AND not unknown)
	// When Computed: true and user passes null, Terraform marks it as unknown, not null.
	// IsUnknown() is true for new resources where the value will be computed.
	// We must skip setting this to avoid sending false to the API for CI/Merge jobs.
	var forceNodeSelection *bool
	if !plan.ForceNodeSelection.IsNull() && !plan.ForceNodeSelection.IsUnknown() {
		fns := plan.ForceNodeSelection.ValueBool()
		forceNodeSelection = &fns
	}

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

	// Extract cost_optimization_features from plan
	var costOptimizationFeatures []string
	if !plan.CostOptimizationFeatures.IsNull() && !plan.CostOptimizationFeatures.IsUnknown() {
		for _, elem := range plan.CostOptimizationFeatures.Elements() {
			if strVal, ok := elem.(types.String); ok && !strVal.IsNull() {
				costOptimizationFeatures = append(costOptimizationFeatures, strVal.ValueString())
			}
		}
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
		forceNodeSelection,
		costOptimizationFeatures,
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

	// Ensure compare_changes_flags is ALWAYS known after Create.
	// If the API does not return a value, explicitly set it to null (not unknown).
	if createdJob.RunCompareChanges != nil {
		plan.RunCompareChanges = types.BoolValue(*createdJob.RunCompareChanges)
	} else {
		plan.RunCompareChanges = types.BoolValue(false)
	}
	if createdJob.CompareChangesFlags != nil {
		plan.CompareChangesFlags = types.StringValue(*createdJob.CompareChangesFlags)
	} else {
		plan.CompareChangesFlags = types.StringNull()
	}

	// Populate force_node_selection from API response
	if createdJob.ForceNodeSelection != nil {
		plan.ForceNodeSelection = types.BoolValue(*createdJob.ForceNodeSelection)
	} else {
		// If not set in config and API doesn't return it, keep it null
		if plan.ForceNodeSelection.IsNull() {
			plan.ForceNodeSelection = types.BoolNull()
		}
	}

	// Populate cost_optimization_features from API response.
	// Use empty set (not null) so UseStateForUnknown() works on subsequent plans.
	if len(createdJob.CostOptimizationFeatures) > 0 {
		features := make([]attr.Value, len(createdJob.CostOptimizationFeatures))
		for i, f := range createdJob.CostOptimizationFeatures {
			features[i] = types.StringValue(f)
		}
		plan.CostOptimizationFeatures, _ = types.SetValue(types.StringType, features)
	} else {
		plan.CostOptimizationFeatures, _ = types.SetValue(types.StringType, []attr.Value{})
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

	// #region agent log
	debugJobTypeLog("pre-fix", "H4", "pkg/framework/objects/job/resource.go:Read", "Read retrieved job", map[string]any{
		"job_id":                        jobIDStr,
		"retrieved_job_type":            retrievedJob.JobType,
		"retrieved_trigger_schedule":    retrievedJob.Triggers.Schedule,
		"retrieved_trigger_on_merge":    retrievedJob.Triggers.OnMerge,
		"retrieved_trigger_git_provider": retrievedJob.Triggers.GitProviderWebhook,
		"retrieved_trigger_github":      retrievedJob.Triggers.GithubWebhook,
	})
	// #endregion

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
	if retrievedJob.Schedule.Date.Type == "interval_cron" && retrievedJob.Schedule.Date.Cron != nil {
		// For interval_cron, parse the interval from the cron expression (e.g., "4 */5 * * 0,1,2,3,4,5,6")
		cronParts := strings.Split(*retrievedJob.Schedule.Date.Cron, " ")
		if len(cronParts) >= 2 && strings.HasPrefix(cronParts[1], "*/") {
			if intervalVal, err := strconv.Atoi(strings.TrimPrefix(cronParts[1], "*/")); err == nil {
				schedule = intervalVal
			}
		}
	} else if retrievedJob.Schedule.Time.Interval > 0 {
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

	if retrievedJob.RunCompareChanges != nil {
		state.RunCompareChanges = types.BoolValue(*retrievedJob.RunCompareChanges)
	} else {
		state.RunCompareChanges = types.BoolValue(false)
	}
	if retrievedJob.CompareChangesFlags != nil {
		state.CompareChangesFlags = types.StringValue(*retrievedJob.CompareChangesFlags)
	} else {
		state.CompareChangesFlags = types.StringNull()
	}
	state.RunLint = types.BoolValue(retrievedJob.RunLint)
	state.ErrorsOnLintFailure = types.BoolValue(retrievedJob.ErrorsOnLintFailure)

	if retrievedJob.ForceNodeSelection != nil {
		state.ForceNodeSelection = types.BoolValue(*retrievedJob.ForceNodeSelection)
	} else {
		state.ForceNodeSelection = types.BoolNull()
	}

	// Populate cost_optimization_features from API response.
	// Always use an empty set (not null) when no features are returned, so that
	// UseStateForUnknown() has a known value to preserve during plan and avoids
	// perpetual "+ cost_optimization_features = (known after apply)" diffs.
	if len(retrievedJob.CostOptimizationFeatures) > 0 {
		features := make([]attr.Value, len(retrievedJob.CostOptimizationFeatures))
		for i, f := range retrievedJob.CostOptimizationFeatures {
			features[i] = types.StringValue(f)
		}
		state.CostOptimizationFeatures, _ = types.SetValue(types.StringType, features)
	} else {
		state.CostOptimizationFeatures, _ = types.SetValue(types.StringType, []attr.Value{})
	}

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

	planJobType := ""
	if !plan.JobType.IsNull() && !plan.JobType.IsUnknown() {
		planJobType = plan.JobType.ValueString()
	}
	stateJobType := ""
	if !state.JobType.IsNull() && !state.JobType.IsUnknown() {
		stateJobType = state.JobType.ValueString()
	}
	planScheduleTrigger := false
	planOnMergeTrigger := false
	planGitProviderTrigger := false
	planGithubTrigger := false
	if plan.Triggers != nil {
		planScheduleTrigger = plan.Triggers.Schedule.ValueBool()
		planOnMergeTrigger = plan.Triggers.OnMerge.ValueBool()
		planGitProviderTrigger = plan.Triggers.GitProviderWebhook.ValueBool()
		planGithubTrigger = plan.Triggers.GithubWebhook.ValueBool()
	}
	// #region agent log
	debugJobTypeLog("pre-fix", "H1", "pkg/framework/objects/job/resource.go:Update", "Update plan/state before GetJob", map[string]any{
		"job_id":                        state.ID.ValueInt64(),
		"plan_job_type_is_null":         plan.JobType.IsNull(),
		"plan_job_type_is_unknown":      plan.JobType.IsUnknown(),
		"plan_job_type":                 planJobType,
		"state_job_type_is_null":        state.JobType.IsNull(),
		"state_job_type_is_unknown":     state.JobType.IsUnknown(),
		"state_job_type":                stateJobType,
		"plan_trigger_schedule":         planScheduleTrigger,
		"plan_trigger_on_merge":         planOnMergeTrigger,
		"plan_trigger_git_provider":     planGitProviderTrigger,
		"plan_trigger_github_webhook":   planGithubTrigger,
		"plan_schedule_type":            plan.ScheduleType.ValueString(),
	})
	// #endregion

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

	// #region agent log
	debugJobTypeLog("pre-fix", "H2", "pkg/framework/objects/job/resource.go:Update", "Remote job from GetJob before mutation", map[string]any{
		"job_id":                       jobIDStr,
		"remote_before_job_type":       job.JobType,
		"remote_trigger_schedule":      job.Triggers.Schedule,
		"remote_trigger_on_merge":      job.Triggers.OnMerge,
		"remote_trigger_git_provider":  job.Triggers.GitProviderWebhook,
		"remote_trigger_github_webhook": job.Triggers.GithubWebhook,
		"remote_schedule_type":         job.Schedule.Date.Type,
	})
	// #endregion

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
	if scheduleType == "interval_cron" {
		// For interval_cron, build the cron expression like CreateJob does
		daysStr := make([]string, len(plan.ScheduleDays))
		for i, day := range plan.ScheduleDays {
			daysStr[i] = strconv.Itoa(int(day.ValueInt64()))
		}
		cronExpr := fmt.Sprintf("4 */%d * * %s", scheduleInterval, strings.Join(daysStr, ","))
		job.Schedule.Date.Cron = &cronExpr
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

	job.RunLint = plan.RunLint.ValueBool()
	job.ErrorsOnLintFailure = plan.ErrorsOnLintFailure.ValueBool()
	// IMPORTANT:
	// We start from the remote job object (GetJob), so we MUST clear SAO fields first.
	// Otherwise, disabling SAO in Terraform would keep sending old SAO pointers back to the API.
	job.RunCompareChanges = nil
	job.CompareChangesFlags = nil

	// Only set RunCompareChanges / CompareChangesFlags when SAO is explicitly enabled in config.
	// Note: compare_changes_flags can be Unknown/Computed in the plan; never read ValueString() unless Known+NonNull.
	if !plan.RunCompareChanges.IsUnknown() && plan.RunCompareChanges.ValueBool() {
		runCompareChanges := true
		job.RunCompareChanges = &runCompareChanges
		if !plan.CompareChangesFlags.IsUnknown() && !plan.CompareChangesFlags.IsNull() {
			ccf := plan.CompareChangesFlags.ValueString()
			job.CompareChangesFlags = &ccf
		}
	}

	// ForceNodeSelection: Only set if explicitly provided (not null AND not unknown)
	// When unknown, let the API determine the value (important for CI/Merge jobs)
	if plan.ForceNodeSelection.IsNull() || plan.ForceNodeSelection.IsUnknown() {
		job.ForceNodeSelection = nil
	} else {
		fns := plan.ForceNodeSelection.ValueBool()
		job.ForceNodeSelection = &fns
	}

	// Handle cost_optimization_features updates
	if !plan.CostOptimizationFeatures.IsNull() && !plan.CostOptimizationFeatures.IsUnknown() {
		var features []string
		for _, elem := range plan.CostOptimizationFeatures.Elements() {
			if strVal, ok := elem.(types.String); ok && !strVal.IsNull() {
				features = append(features, strVal.ValueString())
			}
		}
		job.CostOptimizationFeatures = features
	} else {
		job.CostOptimizationFeatures = nil
	}

	// Handle job_type updates with validation
	// Only validate and set if the plan has an explicit non-empty job_type value
	if !plan.JobType.IsNull() && plan.JobType.ValueString() != "" {
		newJobType := plan.JobType.ValueString()
		prevJobType := job.JobType // This is the current value from the API

		// Validate the job type change
		if err := validateJobTypeChange(prevJobType, newJobType); err != nil {
			resp.Diagnostics.AddError(
				"Invalid job_type change",
				fmt.Sprintf("Cannot change job_type from '%s' to '%s': %s", prevJobType, newJobType, err.Error()),
			)
			return
		}

		job.JobType = newJobType
	}

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

	// #region agent log
	debugJobTypeLog("pre-fix", "H3", "pkg/framework/objects/job/resource.go:Update", "UpdateJob response used for new state", map[string]any{
		"job_id":                         jobIDStr,
		"update_response_job_type":       updatedJob.JobType,
		"update_response_trigger_schedule": updatedJob.Triggers.Schedule,
		"update_response_trigger_on_merge": updatedJob.Triggers.OnMerge,
		"update_response_trigger_git_provider": updatedJob.Triggers.GitProviderWebhook,
		"update_response_trigger_github_webhook": updatedJob.Triggers.GithubWebhook,
		"update_response_schedule_type":  updatedJob.Schedule.Date.Type,
	})
	// #endregion

	if updatedJob.JobType != "" {
		plan.JobType = types.StringValue(updatedJob.JobType)
	} else {
		plan.JobType = types.StringNull()
	}

	// Ensure run_compare_changes / compare_changes_flags are ALWAYS known after Update.
	// If the API does not return values, explicitly set to false/null (not unknown).
	if updatedJob.RunCompareChanges != nil {
		plan.RunCompareChanges = types.BoolValue(*updatedJob.RunCompareChanges)
	} else {
		plan.RunCompareChanges = types.BoolValue(false)
	}
	if updatedJob.CompareChangesFlags != nil {
		plan.CompareChangesFlags = types.StringValue(*updatedJob.CompareChangesFlags)
	} else {
		plan.CompareChangesFlags = types.StringNull()
	}

	// Populate force_node_selection from API response
	if updatedJob.ForceNodeSelection != nil {
		plan.ForceNodeSelection = types.BoolValue(*updatedJob.ForceNodeSelection)
	} else {
		plan.ForceNodeSelection = types.BoolNull()
	}

	// Populate cost_optimization_features from API response.
	// Use empty set (not null) so UseStateForUnknown() works on subsequent plans.
	if len(updatedJob.CostOptimizationFeatures) > 0 {
		features := make([]attr.Value, len(updatedJob.CostOptimizationFeatures))
		for i, f := range updatedJob.CostOptimizationFeatures {
			features[i] = types.StringValue(f)
		}
		plan.CostOptimizationFeatures, _ = types.SetValue(types.StringType, features)
	} else {
		plan.CostOptimizationFeatures, _ = types.SetValue(types.StringType, []attr.Value{})
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

func (j *jobResource) validateExecuteSteps(executeSteps []string) error {
	dbt_flags := []string{
		"--warn-error",
		"--use-experimental-parser",
		"--no-partial-parse",
		"--fail-fast",
	}

	dbt_commands := []string{
		"run",
		"test",
		"archive",
		"snapshot",
		"seed",
		"source",
		"compile",
		"ls",
		"list",
		`docs\s+generate`,
		"parse",
		"build",
		"clone",
		"debug",
		"retry",
		"compare",
		"sl",
	}

	// Build regex pattern for valid commands
	flagsPattern := strings.Join(dbt_flags, "|")
	commandsPattern := strings.Join(dbt_commands, "|")
	validCommandsPattern := fmt.Sprintf(`^\s*dbt\s+((%s)\s+)*(%s)\s*.*$`, flagsPattern, commandsPattern)
	validCommandsRegex := regexp.MustCompile(validCommandsPattern)

	// Validate each execute step individually
	for _, step := range executeSteps {
		// Check if step matches valid dbt command pattern
		if !validCommandsRegex.MatchString(step) {
			return fmt.Errorf("invalid command: %s. Allowed commands are: %s", step, strings.Join(dbt_commands, ", "))
		}

		// Check that each flag isn't used more than once within this step
		for _, flag := range dbt_flags {
			if strings.Count(step, flag) > 1 {
				return fmt.Errorf("flag %s can only be used once per step in: %s", flag, step)
			}
		}
	}

	return nil
}
