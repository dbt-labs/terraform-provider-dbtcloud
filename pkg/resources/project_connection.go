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

var projectConnectionSchema = map[string]*schema.Schema{
	"connection_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Connection ID",
	},
	"project_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID",
	},
}

func ResourceProjectConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectConnectionCreate,
		ReadContext:   resourceProjectConnectionRead,
		UpdateContext: resourceProjectConnectionUpdate,
		DeleteContext: resourceProjectConnectionDelete,

		CustomizeDiff: customdiff.All(
			customdiff.ForceNewIfChange("connection_id", func(_ context.Context, old, new, meta interface{}) bool { return true }),
			customdiff.ForceNewIfChange("project_id", func(_ context.Context, old, new, meta interface{}) bool { return true }),
		),

		Schema: projectConnectionSchema,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceProjectConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	connectionID := d.Get("connection_id").(int)
	projectID := d.Get("project_id").(int)
	projectIDString := strconv.Itoa(projectID)

	project, err := c.GetProject(projectIDString)
	if err != nil {
		return diag.FromErr(err)
	}

	project.ConnectionID = &connectionID

	_, err = c.UpdateProject(projectIDString, *project)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", *project.ID, dbt_cloud.ID_DELIMITER, project.ConnectionID))

	resourceProjectConnectionRead(ctx, d, m)

	return diags
}

func resourceProjectConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	if err := d.Set("connection_id", project.ConnectionID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", project.ID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceProjectConnectionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	return resourceProjectConnectionRead(ctx, d, m)
}

func resourceProjectConnectionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID := d.Get("project_id").(int)
	projectIDString := strconv.Itoa(projectID)

	project, err := c.GetProject(projectIDString)
	if err != nil {
		return diag.FromErr(err)
	}

	project.ConnectionID = nil

	_, err = c.UpdateProject(projectIDString, *project)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
