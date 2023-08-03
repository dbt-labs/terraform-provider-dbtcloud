package data_sources

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var environmentSchema = map[string]*schema.Schema{
	"environment_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "ID of the environment",
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
	"credential_id": &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Credential ID to create the environment with",
	},
	"name": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Environment name",
	},
	"dbt_version": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Version number of dbt to use in this environment, usually in the format 1.2.0-latest rather than core versions",
	},
	"type": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The type of environment (must be either development or deployment)",
	},
	"use_custom_branch": &schema.Schema{
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether to use a custom git branch in this environment",
	},
	"custom_branch": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Which custom branch to use in this environment",
	},
	"deployment_type": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The type of deployment environment (currently 'production' or empty)",
	},
}

func DatasourceEnvironment() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceEnvironmentRead,
		Schema:      environmentSchema,
	}
}

func datasourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	environmentID := d.Get("environment_id").(int)
	projectID := d.Get("project_id").(int)

	environment, err := c.GetEnvironment(projectID, environmentID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("is_active", environment.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", environment.Project_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("credential_id", environment.Credential_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", environment.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dbt_version", environment.Dbt_Version); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", environment.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("use_custom_branch", environment.Use_Custom_Branch); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("custom_branch", environment.Custom_Branch); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("environment_id", environment.Environment_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("deployment_type", environment.DeploymentType); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", environment.Project_Id, dbt_cloud.ID_DELIMITER, *environment.ID))

	return diags
}
