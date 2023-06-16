package resources

import (
	"context"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var projectSchema = map[string]*schema.Schema{
	"name": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "Project name",
	},
	"dbt_project_subdirectory": &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "dbt project subdirectory path",
	},
}

func ResourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,

		Schema: projectSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID := d.Id()

	project, err := c.GetProject(projectID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", project.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dbt_project_subdirectory", project.DbtProjectSubdirectory); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	dbtProjectSubdirectory := d.Get("dbt_project_subdirectory").(string)

	p, err := c.CreateProject(name, dbtProjectSubdirectory)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(*p.ID))

	resourceProjectRead(ctx, d, m)

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)
	projectID := d.Id()

	if d.HasChange("name") || d.HasChange("dbt_project_subdirectory") {
		project, err := c.GetProject(projectID)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("name") {
			name := d.Get("name").(string)
			project.Name = name
		}
		if d.HasChange("dbt_project_subdirectory") {
			dbtProjectSubdirectory := d.Get("dbt_project_subdirectory").(string)
			project.DbtProjectSubdirectory = &dbtProjectSubdirectory
		}

		_, err = c.UpdateProject(projectID, *project)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceProjectRead(ctx, d, m)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)
	projectID := d.Id()

	var diags diag.Diagnostics

	project, err := c.GetProject(projectID)
	if err != nil {
		return diag.FromErr(err)
	}

	project.State = dbt_cloud.STATE_DELETED
	_, err = c.UpdateProject(projectID, *project)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
