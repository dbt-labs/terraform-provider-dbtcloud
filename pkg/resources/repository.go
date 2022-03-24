package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceRepository() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRepositoryCreate,
		ReadContext:   resourceRepositoryRead,
		UpdateContext: resourceRepositoryUpdate,
		DeleteContext: resourceRepositoryDelete,

		Schema: map[string]*schema.Schema{
			"repository_id": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Repository Identifier",
			},
			"is_active": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the repository is active",
			},
			"project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Project ID to create the environment in",
			},
			"remote_url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Git URL for the repository or <Group>/<Project> for Gitlab",
			},
			"git_clone_strategy": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "deploy_key",
				Description: "Git clone strategy for the repository",
			},
			"repository_credentials_id": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Credentials ID for the repository (From the repository side not the DBT Cloud ID)",
			},
			"gitlab_project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Identifier for the Gitlab project",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceRepositoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	isActive := d.Get("is_active").(bool)
	projectId := d.Get("project_id").(int)
	remoteUrl := d.Get("remote_url").(string)
	gitCloneStrategy := d.Get("git_clone_strategy").(string)
	repositoryCredentialsID := d.Get("repository_credentials_id").(int)
	gitlabProjectID := d.Get("gitlab_project_id").(int)

	repository, err := c.CreateRepository(projectId, remoteUrl, isActive, gitCloneStrategy, repositoryCredentialsID, gitlabProjectID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", repository.ProjectID, dbt_cloud.ID_DELIMITER, *repository.ID))

	resourceRepositoryRead(ctx, d, m)

	return diags
}

func resourceRepositoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	repositoryIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	repository, err := c.GetRepository(repositoryIdString, projectIdString)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("repository_id", repository.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_active", repository.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", repository.ProjectID); err != nil {
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

	return diags
}

func resourceRepositoryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	repositoryIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	if d.HasChange("remote_url") || d.HasChange("is_active") || d.HasChange("git_clone_strategy") || d.HasChange("repository_credentials_id") {
		repository, err := c.GetRepository(repositoryIdString, projectIdString)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("remote_url") {
			remoteUrl := d.Get("remote_url").(string)
			repository.RemoteUrl = remoteUrl
		}
		if d.HasChange("is_active") {
			isActive := d.Get("is_active").(bool)
			if isActive {
				repository.State = dbt_cloud.STATE_ACTIVE
			} else {
				repository.State = dbt_cloud.STATE_DELETED
			}
		}
		if d.HasChange("git_clone_strategy") {
			gitCloneStrategy := d.Get("git_clone_strategy").(string)
			repository.GitCloneStrategy = gitCloneStrategy
		}
		if d.HasChange("repository_credentials_id") {
			repositoryCredentialsID := d.Get("repository_credentials_id").(int)
			repository.RepositoryCredentialsID = &repositoryCredentialsID
		}

		_, err = c.UpdateRepository(repositoryIdString, projectIdString, *repository)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceRepositoryRead(ctx, d, m)
}

func resourceRepositoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	repositoryIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	_, err := c.DeleteRepository(repositoryIdString, projectIdString)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
