package resources

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ENVIRONMENT_STATE_ACTIVE = 1
const ENVIRONMENT_STATE_DELETED = 2

func ResourceEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvironmentCreate,
		ReadContext:   resourceEnvironmentRead,
		UpdateContext: resourceEnvironmentUpdate,
		DeleteContext: resourceEnvironmentDelete,

		Schema: map[string]*schema.Schema{
			"is_active": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the environment is active",
			},
			"project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Project ID to create the environment in",
			},
			"credential_id": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     nil,
				Description: "Credential ID to create the environment with",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Environment name",
			},
			"dbt_version": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Version number of dbt to use in this environment",
			},
			"type": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The type of environment (must be either development or deployment)",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					type_ := val.(string)
					switch type_ {
					case
						"development",
						"deployment":
						return
					}
					errs = append(errs, fmt.Errorf("%q must be either development or deployment, got: %q", key, type_))
					return
				},
			},
			"use_custom_branch": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to use a custom git branch in this environment",
			},
			"custom_branch": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Which custom branch to use in this environment",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	isActive := d.Get("is_active").(bool)
	projectId := d.Get("project_id").(int)
	credentialId := d.Get("credential_id").(int)
	name := d.Get("name").(string)
	dbtVersion := d.Get("dbt_version").(string)
	type_ := d.Get("type").(string)
	useCustomBranch := d.Get("use_custom_branch").(bool)
	customBranch := d.Get("custom_branch").(string)

	environment, err := c.CreateEnvironment(isActive, projectId, name, dbtVersion, type_, useCustomBranch, customBranch, credentialId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", environment.Project_Id, dbt_cloud.ID_DELIMITER, *environment.ID))

	resourceJobRead(ctx, d, m)

	return diags
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	environmentId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	environment, err := c.GetEnvironment(projectId, environmentId)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("is_active", environment.State == ENVIRONMENT_STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", environment.Project_Id); err != nil {
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

	return diags
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectId, err := strconv.Atoi(strings.Split(d.Id(), ",")[0])
	if err != nil {
		return diag.FromErr(err)
	}

	environmentId, err := strconv.Atoi(strings.Split(d.Id(), ",")[1])
	if err != nil {
		return diag.FromErr(err)
	}

	// TODO: add more changes here

	if d.HasChange("name") {
		environment, err := c.GetEnvironment(projectId, environmentId)
		if err != nil {
			return diag.FromErr(err)
		}

		name := d.Get("name").(string)
		environment.Name = name
		_, err = c.UpdateEnvironment(projectId, environmentId, *environment)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceEnvironmentRead(ctx, d, m)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	environmentId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("Environment deleting is not yet supported in dbt Cloud, setting state to deleted")

	var diags diag.Diagnostics

	environment, err := c.GetEnvironment(projectId, environmentId)
	if err != nil {
		return diag.FromErr(err)
	}

	environment.State = ENVIRONMENT_STATE_DELETED
	_, err = c.UpdateEnvironment(projectId, environmentId, *environment)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
