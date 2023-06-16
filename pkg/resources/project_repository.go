package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var projectRepositorySchema = map[string]*schema.Schema{
	"repository_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Repository ID",
	},
	"project_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID",
	},
}

func ResourceProjectRepository() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectRepositoryCreate,
		ReadContext:   resourceProjectRepositoryRead,
		UpdateContext: resourceProjectRepositoryUpdate,
		DeleteContext: resourceProjectRepositoryDelete,

		CustomizeDiff: customdiff.All(
			customdiff.ForceNewIfChange("repository_id", func(_ context.Context, old, new, meta interface{}) bool { return true }),
			customdiff.ForceNewIfChange("project_id", func(_ context.Context, old, new, meta interface{}) bool { return true }),
		),

		Schema: projectRepositorySchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceProjectRepositoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceProjectRepositoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}
	projectIDString := strconv.Itoa(projectID)

	project, err := c.GetProject(projectIDString)
	if err != nil {
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

func resourceProjectRepositoryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	return resourceProjectRepositoryRead(ctx, d, m)
}

func resourceProjectRepositoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
