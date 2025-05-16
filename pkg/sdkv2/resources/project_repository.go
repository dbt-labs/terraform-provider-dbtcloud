package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var projectRepositorySchema = map[string]*schema.Schema{
	"repository_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Repository ID",
		ForceNew:    true,
	},
	"project_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID",
		ForceNew:    true,
	},
}

func ResourceProjectRepository() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectRepositoryCreate,
		ReadContext:   resourceProjectRepositoryRead,
		DeleteContext: resourceProjectRepositoryDelete,

		Schema: projectRepositorySchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "This resource allows you to link a dbt Cloud project to a git repository.",
	}
}

func resourceProjectRepositoryCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	repositoryID := d.Get("repository_id").(int)
	projectID := d.Get("project_id").(int)
	projectIDString := strconv.Itoa(projectID)

	project, err := c.GetProject(projectIDString)
	if err != nil {
		return diag.FromErr(err)
	}

	project.RepositoryID = &repositoryID

	_, err = c.UpdateProject(projectIDString, *project)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", *project.ID, dbt_cloud.ID_DELIMITER, project.RepositoryID))

	resourceProjectRepositoryRead(ctx, d, m)

	return diags
}

func resourceProjectRepositoryRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics
	projectIDString, _, err := helper.SplitIDToStrings(
		d.Id(),
		"dbtcloud_project_repository",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	project, err := c.GetProject(projectIDString)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("repository_id", project.RepositoryID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", project.ID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceProjectRepositoryDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID := d.Get("project_id").(int)
	projectIDString := strconv.Itoa(projectID)

	project, err := c.GetProject(projectIDString)
	if err != nil {
		return diag.FromErr(err)
	}

	project.RepositoryID = nil

	_, err = c.UpdateProject(projectIDString, *project)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
