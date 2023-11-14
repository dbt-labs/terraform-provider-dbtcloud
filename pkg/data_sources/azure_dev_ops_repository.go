package data_sources

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var azureDevOpsRepositorySchema = map[string]*schema.Schema{
	"id": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The internal Azure Dev Ops ID of the ADO Repository",
	},
	"name": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the ADO repository",
	},
	"azure_dev_ops_project_id": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The internal Azure Dev Ops ID of the ADO Project. Can be retrieved using the data source dbtcloud_azure_dev_ops_project and the project name",
	},
	"details_url": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The URL of the ADO repository showing details about the repository and its attributes",
	},
	"remote_url": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The HTTP URL of the ADO repository used to connect to dbt Cloud",
	},
	"web_url": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The URL of the ADO repository accessible in the browser",
	},
	"default_branch": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The default branch of the ADO repository",
	},
}

func DatasourceAzureDevOpsRepository() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAzureDevOpsRepositoryRead,
		Schema:      azureDevOpsRepositorySchema,
		Description: `Use this data source to retrieve the ID and details of an Azure Dev Ops repository 
based on its name and the ID of the Azure Dev Ops project it belongs to.
		
This data source requires connecting with a user token and doesn't work with a service token.`,
	}
}

func datasourceAzureDevOpsRepositoryRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	repositoryName := d.Get("name").(string)
	azureDevOpsProjectID := d.Get("azure_dev_ops_project_id").(string)

	azureDevOpsRepository, err := c.GetAzureDevOpsRepository(repositoryName, azureDevOpsProjectID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("id", azureDevOpsRepository.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", azureDevOpsRepository.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("azure_dev_ops_project_id", azureDevOpsProjectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("details_url", azureDevOpsRepository.DetailsURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("remote_url", azureDevOpsRepository.RemoteURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("web_url", azureDevOpsRepository.WebURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_branch", azureDevOpsRepository.DefaultBranch); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(azureDevOpsRepository.ID)

	return diags
}
