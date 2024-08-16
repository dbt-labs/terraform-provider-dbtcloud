package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var projectConnectionSchema = map[string]*schema.Schema{
	"connection_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Connection ID",
		ForceNew:    true,
	},
	"project_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID",
		ForceNew:    true,
	},
}

func ResourceProjectConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectConnectionCreate,
		ReadContext:   resourceProjectConnectionRead,
		DeleteContext: resourceProjectConnectionDelete,

		Schema:             projectConnectionSchema,
		DeprecationMessage: "This resource is deprecated with the release of global connections and it will be removed in a future version of the provider. Going forward, please set the `connection_id` in the `dbtcloud_environment` resource instead.",
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceProjectConnectionCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
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

func resourceProjectConnectionRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}
	projectIDString := strconv.Itoa(projectID)

	project, err := c.GetProject(projectIDString)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
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

func resourceProjectConnectionDelete(
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

	project.ConnectionID = nil

	_, err = c.UpdateProject(projectIDString, *project)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
