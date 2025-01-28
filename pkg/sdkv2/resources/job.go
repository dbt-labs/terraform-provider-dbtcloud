package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/samber/lo"
)

var (
	scheduleTypes = []string{
		"every_day",
		"days_of_week",
		"custom_cron",
	}
)

var jobSchema = map[string]*schema.Schema{
	"project_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID to create the job in",
		ForceNew:    true,
	},
	"environment_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Environment ID to create the job in",
	},
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Job name",
	},
	"description": {
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "",
		Description: "Description for the job",
	},
	"execute_steps": {
		Type:     schema.TypeList,
		MinItems: 1,
		Required: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Description: "List of commands to execute for the job",
	},
	"dbt_version": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Version number of dbt to use in this job, usually in the format 1.2.0-latest rather than core versions",
	},
	"is_active": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "Should always be set to true as setting it to false is the same as creating a job in a deleted state. To create/keep a job in a 'deactivated' state, check  the `triggers` config.",
	},
	"triggers": {
		Type:     schema.TypeMap,
		Required: true,
		Elem: &schema.Schema{
			Type:     schema.TypeBool,
			Optional: false,
			Default:  false,
		},
		Description: "Flags for which types of triggers to use, the values are `github_webhook`, `git_provider_webhook`, `schedule` and `on_merge`. All flags should be listed and set with `true` or `false`. When `on_merge` is `true`, all the other values must be false.<br>`custom_branch_only` used to be allowed but has been deprecated from the API. The jobs will use the custom branch of the environment. Please remove the `custom_branch_only` from your config. <br>To create a job in a 'deactivated' state, set all to `false`.",
	},
	"num_threads": {
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     1,
		Description: "Number of threads to use in the job",
	},
	"target_name": {
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "default",
		Description: "Target name for the dbt profile",
	},
	"generate_docs": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Flag for whether the job should generate documentation",
	},
	"run_generate_sources": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Flag for whether the job should add a `dbt source freshness` step to the job. The difference between manually adding a step with `dbt source freshness` in the job steps or using this flag is that with this flag, a failed freshness will still allow the following steps to run.",
	},
	"run_lint": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Whether the CI job should lint SQL changes. Defaults to `false`.",
	},
	"errors_on_lint_failure": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "Whether the CI job should fail when a lint error is found. Only used when `run_lint` is set to `true`. Defaults to `true`.",
	},
	"schedule_type": {
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "every_day",
		Description:  "Type of schedule to use, one of every_day/ days_of_week/ custom_cron",
		ValidateFunc: validation.StringInSlice(scheduleTypes, false),
	},
	"schedule_interval": {
		Type:          schema.TypeInt,
		Optional:      true,
		Default:       1,
		Description:   "Number of hours between job executions if running on a schedule",
		ValidateFunc:  validation.IntBetween(1, 23),
		ConflictsWith: []string{"schedule_hours", "schedule_cron"},
	},
	"schedule_hours": {
		Type:     schema.TypeList,
		MinItems: 1,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
		Description:   "List of hours to execute the job at if running on a schedule",
		ConflictsWith: []string{"schedule_interval", "schedule_cron"},
	},
	"schedule_days": {
		Type:     schema.TypeList,
		MinItems: 1,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
		Description: "List of days of week as numbers (0 = Sunday, 7 = Saturday) to execute the job at if running on a schedule",
	},
	"schedule_cron": {
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "Custom cron expression for schedule",
		ConflictsWith: []string{"schedule_interval", "schedule_hours"},
	},
	"deferring_job_id": {
		Type:          schema.TypeInt,
		Optional:      true,
		Description:   "Job identifier that this job defers to (legacy deferring approach)",
		ConflictsWith: []string{"self_deferring", "deferring_environment_id"},
	},
	"deferring_environment_id": {
		Type:          schema.TypeInt,
		Optional:      true,
		Description:   "Environment identifier that this job defers to (new deferring approach)",
		ConflictsWith: []string{"self_deferring", "deferring_job_id"},
	},
	"self_deferring": {
		Type:          schema.TypeBool,
		Optional:      true,
		Description:   "Whether this job defers on a previous run of itself",
		ConflictsWith: []string{"deferring_job_id"},
	},
	"timeout_seconds": {
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     0,
		Description: "Number of seconds to allow the job to run before timing out",
	},
	"triggers_on_draft_pr": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Whether the CI job should be automatically triggered on draft PRs",
	},
	"job_completion_trigger_condition": {
		Type:     schema.TypeSet,
		Optional: true,
		// using  a set or a list with 1 item is the way in the SDKv2 to define nested objects
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"job_id": {
					Type:        schema.TypeInt,
					Required:    true,
					Description: "The ID of the job that would trigger this job after completion.",
				},
				"project_id": {
					Type:        schema.TypeInt,
					Required:    true,
					Description: "The ID of the project where the trigger job is running in.",
				},
				"statuses": {
					Type:        schema.TypeSet,
					Required:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Description: "List of statuses to trigger the job on. Possible values are `success`, `error` and `canceled`.",
				},
			},
		},
		Description: "Which other job should trigger this job when it finishes, and on which conditions (sometimes referred as 'job chaining').",
	},
	"run_compare_changes": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
		// Once on the plugin framework, put a validation to check that `deferring_environment_id` is set
		Description: "Whether the CI job should compare data changes introduced by the code changes. Requires `deferring_environment_id` to be set. (Advanced CI needs to be activated in the dbt Cloud Account Settings first as well)",
	},
}

func ResourceJob() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceJobCreate,
		ReadContext:   resourceJobRead,
		UpdateContext: resourceJobUpdate,
		DeleteContext: resourceJobDelete,

		Schema: jobSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.All(
			// if we change the job type (CI, merge or "empty"), we need to recreate the job as dbt Cloud doesn't allow updating them
			// the job type is determined by the triggers

			func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
				oldValue, newValue := d.GetChange("triggers")

				oldValueMap := oldValue.(map[string]interface{})
				newValueMap := newValue.(map[string]interface{})

				oldCI, _ := oldValueMap["github_webhook"].(bool)
				oldOnMerge, _ := oldValueMap["on_merge"].(bool)

				oldType := ""
				if oldCI {
					oldType = "ci"
				} else if oldOnMerge {
					oldType = "merge"
				}

				newCI, _ := newValueMap["github_webhook"].(bool)
				newOnMerge, _ := newValueMap["on_merge"].(bool)

				newType := ""
				if newCI {
					newType = "ci"
				} else if newOnMerge {
					newType = "merge"
				}

				if oldType != newType {
					d.ForceNew("triggers")
				}
				return nil
			},
		),
	}
}

func resourceJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	jobId := d.Id()

	job, err := c.GetJob(jobId)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("project_id", job.ProjectId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("environment_id", job.EnvironmentId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", job.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", job.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("execute_steps", job.ExecuteSteps); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dbt_version", job.DbtVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_active", job.State == 1); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("num_threads", job.Settings.Threads); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("target_name", job.Settings.TargetName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("generate_docs", job.GenerateDocs); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("run_generate_sources", job.RunGenerateSources); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("schedule_type", job.Schedule.Date.Type); err != nil {
		return diag.FromErr(err)
	}

	schedule := 1
	if job.Schedule.Time.Interval > 0 {
		schedule = job.Schedule.Time.Interval
	}
	if err := d.Set("schedule_interval", schedule); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schedule_hours", job.Schedule.Time.Hours); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("schedule_days", job.Schedule.Date.Days); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("schedule_cron", job.Schedule.Date.Cron); err != nil {
		return diag.FromErr(err)
	}
	selfDeferring := job.DeferringJobId != nil && strconv.Itoa(*job.DeferringJobId) == jobId
	if !selfDeferring {
		if err := d.Set("deferring_job_id", job.DeferringJobId); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("deferring_environment_id", job.DeferringEnvironmentId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("self_deferring", selfDeferring); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("timeout_seconds", job.Execution.TimeoutSeconds); err != nil {
		return diag.FromErr(err)
	}

	var triggers map[string]interface{}
	triggersInput, _ := json.Marshal(job.Triggers)
	json.Unmarshal(triggersInput, &triggers)

	// for now, we allow people to keep the triggers.custom_branch_only config even if the parameter was deprecated in the API
	// we set the state to the current config value, so it doesn't do anything
	listedTriggers := d.Get("triggers").(map[string]interface{})
	listedCustomBranchOnly, ok := listedTriggers["custom_branch_only"].(bool)
	if ok {
		triggers["custom_branch_only"] = listedCustomBranchOnly
	}

	// we remove triggers.on_merge if it is not set in the config and it is set to false in the remote
	// that way it works if people don't define it, but also works to import jobs that have it set to true
	// TODO: remove this when we make on_merge mandatory
	_, ok = listedTriggers["on_merge"].(bool)
	noOnMergeConfig := !ok
	onMergeRemoteVal, _ := triggers["on_merge"].(bool)
	onMergeRemoteFalse := !onMergeRemoteVal
	if noOnMergeConfig && onMergeRemoteFalse {
		delete(triggers, "on_merge")
	}

	if err := d.Set("triggers", triggers); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("triggers_on_draft_pr", job.TriggersOnDraftPR); err != nil {
		return diag.FromErr(err)
	}

	if job.JobCompletionTrigger == nil {
		if err := d.Set("job_completion_trigger_condition", nil); err != nil {
			return diag.FromErr(err)
		}
	} else {
		triggerCondition := job.JobCompletionTrigger.Condition
		statusesNames := lo.Map(triggerCondition.Statuses, func(status int, idx int) any {
			return utils.JobCompletionTriggerConditionsMappingCodeHuman[status]
		})
		triggerConditionMap := map[string]any{
			"job_id":     triggerCondition.JobID,
			"project_id": triggerCondition.ProjectID,
			"statuses":   statusesNames,
		}
		triggerConditionSet := utils.JobConditionMapToSet(triggerConditionMap)

		if err := d.Set("job_completion_trigger_condition", triggerConditionSet); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("run_compare_changes", job.RunCompareChanges); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("run_lint", job.RunLint); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("errors_on_lint_failure", job.ErrorsOnLintFailure); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceJobCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectId := d.Get("project_id").(int)
	environmentId := d.Get("environment_id").(int)
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	executeSteps := d.Get("execute_steps").([]interface{})
	dbtVersion := d.Get("dbt_version").(string)
	isActive := d.Get("is_active").(bool)
	triggers := d.Get("triggers").(map[string]interface{})
	numThreads := d.Get("num_threads").(int)
	targetName := d.Get("target_name").(string)
	generateDocs := d.Get("generate_docs").(bool)
	runGenerateSources := d.Get("run_generate_sources").(bool)
	scheduleType := d.Get("schedule_type").(string)
	scheduleInterval := d.Get("schedule_interval").(int)
	scheduleHours := d.Get("schedule_hours").([]interface{})
	scheduleDays := d.Get("schedule_days").([]interface{})
	scheduleCron := d.Get("schedule_cron").(string)
	deferringJobId := d.Get("deferring_job_id").(int)
	deferringEnvironmentID := d.Get("deferring_environment_id").(int)
	selfDeferring := d.Get("self_deferring").(bool)
	timeoutSeconds := d.Get("timeout_seconds").(int)
	triggersOnDraftPR := d.Get("triggers_on_draft_pr").(bool)
	runCompareChanges := d.Get("run_compare_changes").(bool)
	runLint := d.Get("run_lint").(bool)
	errorsOnLintFailure := d.Get("errors_on_lint_failure").(bool)

	var jobCompletionTrigger map[string]any
	empty, completionJobID, completionProjectID, completionStatuses := utils.ExtractJobConditionSet(
		d,
	)
	if !empty {
		jobCompletionTrigger = map[string]any{
			"job_id":     completionJobID,
			"project_id": completionProjectID,
			"statuses":   completionStatuses,
		}
	}

	steps := []string{}
	for _, step := range executeSteps {
		steps = append(steps, step.(string))
	}
	hours := []int{}
	for _, hour := range scheduleHours {
		hours = append(hours, hour.(int))
	}
	days := []int{}
	for _, day := range scheduleDays {
		days = append(days, day.(int))
	}

	j, err := c.CreateJob(
		projectId,
		environmentId,
		name,
		description,
		steps,
		dbtVersion,
		isActive,
		triggers,
		numThreads,
		targetName,
		generateDocs,
		runGenerateSources,
		scheduleType,
		scheduleInterval,
		hours,
		days,
		scheduleCron,
		deferringJobId,
		deferringEnvironmentID,
		selfDeferring,
		timeoutSeconds,
		triggersOnDraftPR,
		jobCompletionTrigger,
		runCompareChanges,
		runLint,
		errorsOnLintFailure,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(*j.ID))

	resourceJobRead(ctx, d, m)

	return diags
}

func resourceJobUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)
	jobId := d.Id()

	if d.HasChange("project_id") ||
		d.HasChange("environment_id") ||
		d.HasChange("name") ||
		d.HasChange("description") ||
		d.HasChange("dbt_version") ||
		d.HasChange("num_threads") ||
		d.HasChange("target_name") ||
		d.HasChange("execute_steps") ||
		d.HasChange("run_generate_sources") ||
		d.HasChange("generate_docs") ||
		d.HasChange("triggers") ||
		d.HasChange("schedule_type") ||
		d.HasChange("schedule_interval") ||
		d.HasChange("schedule_hours") ||
		d.HasChange("schedule_days") ||
		d.HasChange("schedule_cron") ||
		d.HasChange("deferring_job_id") ||
		d.HasChange("deferring_environment_id") ||
		d.HasChange("self_deferring") ||
		d.HasChange("timeout_seconds") ||
		d.HasChange("triggers_on_draft_pr") ||
		d.HasChange("job_completion_trigger_condition") ||
		d.HasChange("run_compare_changes") ||
		d.HasChange("run_lint") ||
		d.HasChange("errors_on_lint_failure") {
		job, err := c.GetJob(jobId)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("project_id") {
			projectID := d.Get("project_id").(int)
			job.ProjectId = projectID
		}
		if d.HasChange("environment_id") {
			envID := d.Get("environment_id").(int)
			job.EnvironmentId = envID
		}
		if d.HasChange("name") {
			name := d.Get("name").(string)
			job.Name = name
		}
		if d.HasChange("description") {
			description := d.Get("description").(string)
			job.Description = description
		}
		if d.HasChange("dbt_version") {
			dbtVersion := d.Get("dbt_version").(string)
			job.DbtVersion = &dbtVersion
		}
		if d.HasChange("num_threads") {
			numThreads := d.Get("num_threads").(int)
			job.Settings.Threads = numThreads
		}
		if d.HasChange("target_name") {
			targetName := d.Get("target_name").(string)
			job.Settings.TargetName = targetName
		}
		if d.HasChange("run_generate_sources") {
			runGenerateSources := d.Get("run_generate_sources").(bool)
			job.RunGenerateSources = runGenerateSources
		}
		if d.HasChange("generate_docs") {
			generateDocs := d.Get("generate_docs").(bool)
			job.GenerateDocs = generateDocs
		}
		if d.HasChange("execute_steps") {
			executeSteps := make([]string, len(d.Get("execute_steps").([]interface{})))
			for i, step := range d.Get("execute_steps").([]interface{}) {
				executeSteps[i] = step.(string)
			}
			job.ExecuteSteps = executeSteps
		}
		if d.HasChange("triggers") {
			var ok bool
			newTriggers := d.Get("triggers").(map[string]interface{})
			job.Triggers.GithubWebhook, ok = newTriggers["github_webhook"].(bool)
			if !ok {
				return diag.FromErr(fmt.Errorf("github_webhook was not provided"))
			}
			job.Triggers.GitProviderWebhook, ok = newTriggers["git_provider_webhook"].(bool)
			if !ok {
				return diag.FromErr(fmt.Errorf("git_provider_webhook was not provided"))
			}
			job.Triggers.Schedule, ok = newTriggers["schedule"].(bool)
			if !ok {
				return diag.FromErr(fmt.Errorf("schedule was not provided"))
			}
			job.Triggers.OnMerge, _ = newTriggers["on_merge"].(bool)
			// should work without on_merge for now, so we don't need to test if it was provided
		}

		if d.HasChange("schedule_interval") {
			scheduleInterval := d.Get("schedule_interval").(int)
			job.Schedule.Time.Interval = scheduleInterval
		}
		if d.HasChange("schedule_hours") {

			scheduleHours := make([]int, len(d.Get("schedule_hours").([]interface{})))
			for i, hour := range d.Get("schedule_hours").([]interface{}) {
				scheduleHours[i] = hour.(int)
			}
			if len(d.Get("schedule_hours").([]interface{})) > 0 {
				job.Schedule.Time.Hours = &scheduleHours
				job.Schedule.Time.Type = "at_exact_hours"
				job.Schedule.Time.Interval = 0
			} else {
				job.Schedule.Time.Hours = nil
				job.Schedule.Time.Interval = d.Get("schedule_interval").(int)
				job.Schedule.Time.Type = "every_hour"
			}
		}
		if d.HasChange("schedule_days") {
			scheduleDays := make([]int, len(d.Get("schedule_days").([]interface{})))
			for i, day := range d.Get("schedule_days").([]interface{}) {
				scheduleDays[i] = day.(int)
			}
			if len(d.Get("schedule_days").([]interface{})) > 0 {
				job.Schedule.Date.Days = &scheduleDays
			}
		}
		if d.HasChange("schedule_cron") {
			scheduleCron := d.Get("schedule_cron").(string)
			job.Schedule.Date.Cron = &scheduleCron
		}

		// we set this after the subfields to remove the fields not matching the schedule type
		// if it was before, some of those fields would be set again
		if d.HasChange("schedule_type") {
			scheduleType := d.Get("schedule_type").(string)
			job.Schedule.Date.Type = scheduleType

			if scheduleType == "days_of_week" || scheduleType == "every_day" {
				job.Schedule.Date.Cron = nil
			}
			if scheduleType == "custom_cron" || scheduleType == "every_day" {
				job.Schedule.Date.Days = nil
			}
		}

		if d.HasChange("deferring_job_id") {
			deferringJobId := d.Get("deferring_job_id").(int)
			if deferringJobId != 0 {
				job.DeferringJobId = &deferringJobId
			} else {
				job.DeferringJobId = nil
			}
		}
		if d.HasChange("deferring_environment_id") {
			deferringEnvironmentId := d.Get("deferring_environment_id").(int)
			if deferringEnvironmentId != 0 {
				job.DeferringEnvironmentId = &deferringEnvironmentId
			} else {
				job.DeferringEnvironmentId = nil
			}
		}
		// If self_deferring has been toggled to true, set deferring_job_id as own ID
		// Otherwise, set it back to what deferring_job_id specifies it to be
		if d.HasChange("self_deferring") {
			if d.Get("self_deferring") == true {
				deferringJobID := *job.ID
				job.DeferringJobId = &deferringJobID
			} else {
				deferringJobId := d.Get("deferring_job_id").(int)
				if deferringJobId != 0 {
					job.DeferringJobId = &deferringJobId
				} else {
					job.DeferringJobId = nil
				}
			}
		}
		if d.HasChange("timeout_seconds") {
			timeoutSeconds := d.Get("timeout_seconds").(int)
			job.Execution.TimeoutSeconds = timeoutSeconds
		}
		if d.HasChange("triggers_on_draft_pr") {
			triggersOnDraftPR := d.Get("triggers_on_draft_pr").(bool)
			job.TriggersOnDraftPR = triggersOnDraftPR
		}
		if d.HasChange("job_completion_trigger_condition") {

			empty, completionJobID, completionProjectID, completionStatuses := utils.ExtractJobConditionSet(
				d,
			)
			if empty {
				job.JobCompletionTrigger = nil
			} else {
				jobCondTrigger := dbt_cloud.JobCompletionTrigger{
					Condition: dbt_cloud.JobCompletionTriggerCondition{
						JobID:     completionJobID,
						ProjectID: completionProjectID,
						Statuses:  completionStatuses,
					},
				}
				job.JobCompletionTrigger = &jobCondTrigger
			}
		}
		if d.HasChange("run_compare_changes") {
			runCompareChanges := d.Get("run_compare_changes").(bool)
			job.RunCompareChanges = runCompareChanges
		}
		if d.HasChange("run_lint") {
			runLint := d.Get("run_lint").(bool)
			job.RunLint = runLint
		}
		if d.HasChange("errors_on_lint_failure") {
			errorsOnLintFailure := d.Get("errors_on_lint_failure").(bool)
			job.ErrorsOnLintFailure = errorsOnLintFailure
		}

		_, err = c.UpdateJob(jobId, *job)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceJobRead(ctx, d, m)
}

func resourceJobDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)
	jobId := d.Id()

	var diags diag.Diagnostics

	job, err := c.GetJob(jobId)
	if err != nil {
		return diag.FromErr(err)
	}

	job.State = dbt_cloud.STATE_DELETED
	_, err = c.UpdateJob(jobId, *job)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
