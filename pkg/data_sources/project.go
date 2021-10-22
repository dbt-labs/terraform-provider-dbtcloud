package data_sources

import (
	"context"
	"strconv"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var projectSchema = map[string]*schema.Schema{
	"project_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "ID of the project to represent",
	},
	"name": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Given name for project",
	},
	"connection_id": &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "ID of the connection associated with the project",
	},
	"repository_id": &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "ID of the repository associated with the project",
	},
	"state": &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Project state should be 1 = active, as 2 = deleted",
	},
}

func DatasourceProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceProjectRead,
		Schema:      projectSchema,
	}
}

func datasourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectId := strconv.Itoa(d.Get("project_id").(int))

	project, err := c.GetProject(projectId)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("project_id", project.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", project.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("connection_id", project.ConnectionID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("repository_id", project.RepositoryID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("state", project.State); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(projectId)

	return diags
}
