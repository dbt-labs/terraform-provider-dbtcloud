package data_sources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var repositorySchema = map[string]*schema.Schema{
	"repository_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "ID for the repository",
	},
	"project_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID to create the environment in",
	},
	"is_active": &schema.Schema{
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the environment is active",
	},
	"remote_url": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Connection name",
	},
}

func DatasourceRepository() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceRepositoryRead,
		Schema:      repositorySchema,
	}
}

func datasourceRepositoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	repositoryID := d.Get("repository_id").(int)
	projectID := d.Get("project_id").(int)

	repository, err := c.GetRepository(strconv.Itoa(repositoryID), strconv.Itoa(projectID))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("is_active", repository.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", repository.ProjectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("repository_id", repository.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("remote_url", repository.RemoteUrl); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", repository.ProjectID, dbt_cloud.ID_DELIMITER, *repository.ID))

	return diags
}
