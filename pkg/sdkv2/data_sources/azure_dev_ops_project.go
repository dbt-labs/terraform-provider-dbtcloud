package data_sources

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var azureDevOpsProjectSchema = map[string]*schema.Schema{
	"id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The internal Azure Dev Ops ID of the ADO Project",
	},
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the ADO project",
	},
	"url": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The URL of the ADO project",
	},
}

func DatasourceAzureDevOpsProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAzureDevOpsProjectRead,
		Schema:      azureDevOpsProjectSchema,
		Description: `Use this data source to retrieve the ID of an Azure Dev Ops project 
based on its name.
		
This data source requires connecting with a user token and doesn't work with a service token.`,
	}
}

func datasourceAzureDevOpsProjectRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectName := d.Get("name").(string)

	azureDevOpsProject, err := c.GetAzureDevOpsProject(projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("id", azureDevOpsProject.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", azureDevOpsProject.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("url", azureDevOpsProject.URL); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(azureDevOpsProject.ID)

	return diags
}
