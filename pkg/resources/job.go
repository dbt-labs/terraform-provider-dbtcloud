package resources

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		Optional: true,
		Elem: &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		Description: "Flags for which types of triggers to use, keys of github_webhook, schedule, custom_branch_only",
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

	steps := []string{}
	for _, step := range executeSteps {
		steps = append(steps, step.(string))
	}

	j, err := c.CreateJob(projectId, environmentId, name, steps, dbtVersion, isActive, triggers, numThreads, targetName, generateDocs, runGenerateSources)
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

	if d.HasChange("name") || d.HasChange("dbt_version") {
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
		if d.HasChange("execute_steps") {
			executeSteps := d.Get("execute_steps").([]string)
			job.Execute_Steps = executeSteps
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
	log.Printf("Job deleting is not yet supported in DBT Cloud, setting state to deleted")

	var diags diag.Diagnostics

	job, err := c.GetJob(jobId)
	if err != nil {
		return diag.FromErr(err)
	}

	job.State = 2
	_, err = c.UpdateJob(jobId, *job)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
