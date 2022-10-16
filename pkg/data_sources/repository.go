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
		Description: "Project ID to create the repository in",
	},
	"is_active": &schema.Schema{
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the repository is active",
	},
	"remote_url": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Connection name",
	},
	"git_clone_strategy": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Git clone strategy for the repository",
	},
	"repository_credentials_id": &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Credentials ID for the repository (From the repository side not the DBT Cloud ID)",
	},
	"gitlab_project_id": &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Identifier for the Gitlab project",
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
	fetchDeployKey := d.Get("fetch_deploy_key").(bool)

	repository, err := c.GetRepository(strconv.Itoa(repositoryID), strconv.Itoa(projectID), fetchDeployKey)
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
	if err := d.Set("git_clone_strategy", repository.GitCloneStrategy); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("repository_credentials_id", repository.RepositoryCredentialsID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("gitlab_project_id", repository.GitlabProjectID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", repository.ProjectID, dbt_cloud.ID_DELIMITER, *repository.ID))

	return diags
}
