package resources

import (
	"context"
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

	if err := d.Set("job", job); err != nil {
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
	executeSteps := d.Get("execute_steps").([]string)

	j, err := c.CreateJob(projectId, environmentId, name, executeSteps)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(j.ID))

	resourceJobRead(ctx, d, m)

	return diags
}

func resourceJobUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceJobRead(ctx, d, m)
}

func resourceJobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	log.Printf("Job deleting is not yet supported in DBT Cloud")

	var diags diag.Diagnostics

	return diags
}
