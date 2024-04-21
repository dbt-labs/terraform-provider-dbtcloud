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

func ResourceEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvironmentVariableCreate,
		ReadContext:   resourceEnvironmentVariableRead,
		DeleteContext: resourceEnvironmentVariableDelete,

		Schema: map[string]*schema.Schema{
			"project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Project for the variable to be created in",
				ForceNew:    true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				// as the name is used as the ID, we need to force a new resource if the name changes
				ForceNew:    true,
				Description: "Name for the variable, must be unique within a project, must be prefixed with 'DBT_'",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if !(strings.HasPrefix(v, "DBT_")) {
						errs = append(
							errs,
							fmt.Errorf("the env var must start with DBT_ , got: %s", v),
						)
					}
					return
				},
			},
			"environment_values": &schema.Schema{
				Type:        schema.TypeMap,
				Required:    true,
				Description: "Map from environment names to respective variable value, a special key `project` should be set for the project default variable value. This field is not set as sensitive so take precautions when using secret environment variables.",
				// since the last change in the API we can't just PUSH the new values, so, we can delete it and then create it again
				ForceNew: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceEnvironmentVariableCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID := d.Get("project_id").(int)
	name := d.Get("name").(string)
	environmentValues := d.Get("environment_values").(map[string]interface{})
	environmentValuesStrings := make(map[string]string)
	for envName, value := range environmentValues {
		environmentValuesStrings[envName] = value.(string)
	}

	environmentVariable, err := c.CreateEnvironmentVariable(
		projectID,
		name,
		environmentValuesStrings,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(
		fmt.Sprintf(
			"%d%s%s",
			environmentVariable.ProjectID,
			dbt_cloud.ID_DELIMITER,
			environmentVariable.Name,
		),
	)

	resourceEnvironmentVariableRead(ctx, d, m)

	return diags
}

func resourceEnvironmentVariableRead(
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

	environmentVariableName := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	environmentVariable, err := c.GetEnvironmentVariable(projectID, environmentVariableName)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", environmentVariable.ProjectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", environmentVariable.Name); err != nil {
		return diag.FromErr(err)
	}

	if !strings.HasPrefix(environmentVariable.Name, "DBT_ENV_SECRET_") {
		if err := d.Set("environment_values", environmentVariable.EnvironmentNameValues); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("environment_values", d.Get("environment_values")); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceEnvironmentVariableDelete(
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

	environmentVariableName := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	_, err = c.DeleteEnvironmentVariable(environmentVariableName, projectID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
