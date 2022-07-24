package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvironmentVariableCreate,
		ReadContext:   resourceEnvironmentVariableRead,
		UpdateContext: resourceEnvironmentVariableUpdate,
		DeleteContext: resourceEnvironmentVariableDelete,

		Schema: map[string]*schema.Schema{
			"project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Project for the variable to be created in",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name for the variable, must be unique within a project, must be prefixed with 'DBT_'",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if !(strings.HasPrefix(v, "DBT_")) {
						errs = append(errs, fmt.Errorf("%q must be between 0 and 10 inclusive, got: %s", key, v))
					}
					return
				},
			},
			"environment_values": &schema.Schema{
				Type:        schema.TypeMap,
				Required:    true,
				Description: "Map from environment names to respective variable value",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceEnvironmentVariableCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID := d.Get("project_id").(int)
	name := d.Get("name").(string)
	environmentValues := d.Get("environment_values").(map[string]interface{})
	environmentValuesStrings := make(map[string]string)
	for envName, value := range environmentValues {
		environmentValuesStrings[envName] = value.(string)
	}

	environmentVariable, err := c.CreateEnvironmentVariable(projectID, name, environmentValuesStrings)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%s", environmentVariable.ProjectID, dbt_cloud.ID_DELIMITER, environmentVariable.Name))

	resourceEnvironmentVariableRead(ctx, d, m)

	return diags
}

func resourceEnvironmentVariableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	environmentVariableName := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	environmentVariable, err := c.GetEnvironmentVariable(projectID, environmentVariableName)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("project_id", environmentVariable.ProjectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", environmentVariable.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("environment_values", environmentVariable.EnvironmentNameValues); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceEnvironmentVariableUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectID, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	environmentVariableName := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[2]
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("environment_values") {
		environmentVariable, err := c.GetEnvironmentVariable(projectID, environmentVariableName)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("environment_values") {
			environmentValues := d.Get("environment_values").(map[string]string)
			environmentVariable.EnvironmentNameValues = environmentValues
		}

		_, err = c.UpdateEnvironmentVariable(projectID, *environmentVariable)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceEnvironmentRead(ctx, d, m)
}

func resourceEnvironmentVariableDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	environmentVariableName := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	_, err = c.DeleteEnvironmentVariable(environmentVariableName, projectID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
