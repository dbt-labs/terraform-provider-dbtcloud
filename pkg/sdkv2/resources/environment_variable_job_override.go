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

func ResourceEnvironmentVariableJobOverride() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvironmentVariableJobOverrideCreate,
		ReadContext:   resourceEnvironmentVariableJobOverrideRead,
		UpdateContext: resourceEnvironmentVariableJobOverrideUpdate,
		DeleteContext: resourceEnvironmentVariableJobOverrideDelete,

		Schema: map[string]*schema.Schema{
			"environment_variable_job_override_id": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the environment variable job override",
			},
			"job_definition_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The job ID for which the environment variable is being overridden",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The environment variable name to override",
			},
			"project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The project ID for which the environment variable is being overridden",
			},
			"raw_value": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The value for the override of the environment variable",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceEnvironmentVariableJobOverrideCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	projectId := d.Get("project_id").(int)
	rawValue := d.Get("raw_value").(string)
	jobDefinitionID := d.Get("job_definition_id").(int)

	environmentVariableJobOverride, err := c.CreateEnvironmentVariableJobOverride(
		projectId,
		name,
		rawValue,
		jobDefinitionID,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(
		fmt.Sprintf(
			"%d%s%d%s%d",
			environmentVariableJobOverride.ProjectID,
			dbt_cloud.ID_DELIMITER,
			environmentVariableJobOverride.JobDefinitionID,
			dbt_cloud.ID_DELIMITER,
			*environmentVariableJobOverride.ID,
		),
	)

	resourceEnvironmentVariableJobOverrideRead(ctx, d, m)

	return diags
}

func resourceEnvironmentVariableJobOverrideRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	jobID, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	envVarOverrideID, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[2])
	if err != nil {
		return diag.FromErr(err)
	}

	envVarOverride, err := c.GetEnvironmentVariableJobOverride(
		projectId,
		jobID,
		envVarOverrideID,
	)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("job_definition_id", envVarOverride.JobDefinitionID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", envVarOverride.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", envVarOverride.ProjectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("raw_value", envVarOverride.RawValue); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("environment_variable_job_override_id", *envVarOverride.ID); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceEnvironmentVariableJobOverrideUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	jobID, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	envVarOverrideID, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[2])
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("raw_value") {
		envVarOverride, err := c.GetEnvironmentVariableJobOverride(
			projectId,
			jobID,
			envVarOverrideID,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		rawValue := d.Get("raw_value").(string)
		envVarOverride.RawValue = rawValue

		_, err = c.UpdateEnvironmentVariableJobOverride(
			projectId,
			envVarOverrideID,
			*envVarOverride,
		)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceEnvironmentVariableJobOverrideRead(ctx, d, m)
}

func resourceEnvironmentVariableJobOverrideDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	envVarOverrideID, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[2])
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.DeleteEnvironmentVariableJobOverride(projectId, envVarOverrideID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
