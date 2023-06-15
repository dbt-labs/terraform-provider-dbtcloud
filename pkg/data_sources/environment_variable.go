package data_sources

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var environmentVariableSchema = map[string]*schema.Schema{
	"project_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID the variable exists in",
	},
	"name": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "Name for the variable",
	},
	"environment_values": &schema.Schema{
		Type:        schema.TypeMap,
		Computed:    true,
		Description: "Map containing the environment variables",
	},
}

func DatasourceEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceEnvironmentVariableRead,
		Schema:      environmentVariableSchema,
	}
}

func datasourceEnvironmentVariableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID := d.Get("project_id").(int)
	name := d.Get("name").(string)

	environmentVariable, err := c.GetEnvironmentVariable(projectID, name)
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

	d.SetId(fmt.Sprintf("%d%s%s", environmentVariable.ProjectID, dbt_cloud.ID_DELIMITER, environmentVariable.Name))

	return diags
}
