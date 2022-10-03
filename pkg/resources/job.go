package resources

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	scheduleTypes = []string{
		"every_day",
		"days_of_week",
		"custom_cron",
	}
)

var jobSchema = map[string]*schema.Schema{
	"project_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID to create the job in",
	},
	"environment_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Environment ID to create the job in",
	},
	"name": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "Job name",
	},
	"execute_steps": &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		Required: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Description: "List of commands to execute for the job",
	},
	"dbt_version": &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Version number of DBT to use in this job",
	},
	"is_active": &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "Flag for whether the job is marked active or deleted",
	},
	"triggers": &schema.Schema{
		Type:     schema.TypeMap,
		Required: true,
		Elem: &schema.Schema{
			Type:     schema.TypeBool,
			Optional: false,
			Default:  false,
		},
		Description: "Flags for which types of triggers to use, keys of github_webhook, git_provider_webhook, schedule, custom_branch_only",
	},
	"num_threads": &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     1,
		Description: "Number of threads to use in the job",
	},
	"target_name": &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "default",
		Description: "Target name for the DBT profile",
	},
	"generate_docs": &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Flag for whether the job should generate documentation",
	},
	"run_generate_sources": &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Flag for whether the job should run generate sources",
	},
	"schedule_type": &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "every_day",
		Description:  "Type of schedule to use, one of every_day/ days_of_week/ custom_cron",
		ValidateFunc: validation.StringInSlice(scheduleTypes, false),
	},
	"schedule_interval": &schema.Schema{
		Type:          schema.TypeInt,
		Optional:      true,
		Default:       1,
		Description:   "Number of hours between job executions if running on a schedule",
		ValidateFunc:  validation.IntBetween(1, 23),
		ConflictsWith: []string{"schedule_hours"},
	},
	"schedule_hours": &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
		Description:   "List of hours to execute the job at if running on a schedule",
		ConflictsWith: []string{"schedule_interval"},
	},
	"schedule_days": &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
		Description: "List of days of week as numbers (0 = Sunday, 7 = Saturday) to execute the job at if running on a schedule",
	},
	"schedule_cron": &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Custom cron expression for schedule",
	},
	"deferring_job_id": &schema.Schema{
		Type:          schema.TypeInt,
		Optional:      true,
		Description:   "Job identifier that this job defers to",
		ConflictsWith: []string{"self_deferring"},
	},
	"self_deferring": &schema.Schema{
		Type:          schema.TypeBool,
		Optional:      true,
		Description:   "Whether this job defers on a previous run of itself",
		ConflictsWith: []string{"deferring_job_id"},
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
	}
}

func resourceJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	jobId := d.Id()

	job, err := c.GetJob(jobId)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("project_id", job.Project_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("environment_id", job.Environment_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", job.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("execute_steps", job.Execute_Steps); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dbt_version", job.Dbt_Version); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_active", job.State == 1); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("num_threads", job.Settings.Threads); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("target_name", job.Settings.Target_Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("generate_docs", job.Generate_Docs); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("run_generate_sources", job.Run_Generate_Sources); err != nil {
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
	if err := d.Set("deferring_job_id", job.Deferring_Job_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("self_deferring", job.Deferring_Job_Id != nil && strconv.Itoa(*job.Deferring_Job_Id) == jobId); err != nil {
		return diag.FromErr(err)
	}

	var triggers map[string]interface{}
	triggersInput, _ := json.Marshal(job.Triggers)
	json.Unmarshal(triggersInput, &triggers)
	if err := d.Set("triggers", triggers); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceJobCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectId := d.Get("project_id").(int)
	environmentId := d.Get("environment_id").(int)
	name := d.Get("name").(string)
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
	selfDeferring := d.Get("self_deferring").(bool)

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

	j, err := c.CreateJob(projectId, environmentId, name, steps, dbtVersion, isActive, triggers, numThreads, targetName, generateDocs, runGenerateSources, scheduleType, scheduleInterval, hours, days, scheduleCron, deferringJobId, selfDeferring)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(*j.ID))

	resourceJobRead(ctx, d, m)

	return diags
}

func resourceJobUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)
	jobId := d.Id()

	if d.HasChange("name") || d.HasChange("dbt_version") || d.HasChange("num_threads") ||
		d.HasChange("target_name") || d.HasChange("execute_steps") || d.HasChange("run_generate_sources") ||
		d.HasChange("generate_docs") || d.HasChange("triggers") || d.HasChange("schedule_type") ||
		d.HasChange("schedule_interval") || d.HasChange("schedule_hours") || d.HasChange("schedule_days") ||
		d.HasChange("schedule_cron") || d.HasChange("deferring_job_id") || d.HasChange("self_deferring") {
		job, err := c.GetJob(jobId)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("name") {
			name := d.Get("name").(string)
			job.Name = name
		}
		if d.HasChange("dbt_version") {
			dbtVersion := d.Get("dbt_version").(string)
			job.Dbt_Version = &dbtVersion
		}
		if d.HasChange("num_threads") {
			numThreads := d.Get("num_threads").(int)
			job.Settings.Threads = numThreads
		}
		if d.HasChange("target_name") {
			targetName := d.Get("target_name").(string)
			job.Settings.Target_Name = targetName
		}
		if d.HasChange("run_generate_sources") {
			runGenerateSources := d.Get("run_generate_sources").(bool)
			job.Run_Generate_Sources = runGenerateSources
		}
		if d.HasChange("generate_docs") {
			generateDocs := d.Get("generate_docs").(bool)
			job.Generate_Docs = generateDocs
		}
		if d.HasChange("execute_steps") {
			executeSteps := make([]string, len(d.Get("execute_steps").([]interface{})))
			for i, step := range d.Get("execute_steps").([]interface{}) {
				executeSteps[i] = step.(string)
			}
			job.Execute_Steps = executeSteps
		}
		if d.HasChange("triggers") {
			newTriggers := d.Get("triggers").(map[string]interface{})
			job.Triggers.Github_Webhook = newTriggers["github_webhook"].(bool)
			job.Triggers.GitProviderWebhook = newTriggers["git_provider_webhook"].(bool)
			job.Triggers.Schedule = newTriggers["schedule"].(bool)
			job.Triggers.Custom_Branch_Only = newTriggers["custom_branch_only"].(bool)
		}
		if d.HasChange("schedule_type") {
			scheduleType := d.Get("schedule_type").(string)
			job.Schedule.Date.Type = scheduleType
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
			job.Schedule.Time.Hours = &scheduleHours
			job.Schedule.Time.Type = "at_exact_hours"
			job.Schedule.Time.Interval = 0
		}
		if d.HasChange("schedule_days") {
			scheduleDays := make([]int, len(d.Get("schedule_days").([]interface{})))
			for i, day := range d.Get("schedule_days").([]interface{}) {
				scheduleDays[i] = day.(int)
			}
			job.Schedule.Date.Days = &scheduleDays
		}
		if d.HasChange("schedule_cron") {
			scheduleCron := d.Get("schedule_cron").(string)
			job.Schedule.Date.Cron = &scheduleCron
		}
		if d.HasChange("deferring_job_id") {
			deferringJobId := d.Get("deferring_job_id").(int)
			job.Deferring_Job_Id = &deferringJobId
		}
		// If self_deferring has been toggled to true, set deferring_job_id as own ID
		// Otherwise, set it back to what deferring_job_id specifies it to be
		if d.HasChange("self_deferring") {
			if d.Get("self_deferring") == true {
				job.Deferring_Job_Id = job.ID
			} else {
				deferringJobId := d.Get("deferring_job_id").(int)
				job.Deferring_Job_Id = &deferringJobId
			}
		}

		_, err = c.UpdateJob(jobId, *job)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceJobRead(ctx, d, m)
}

func resourceJobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
