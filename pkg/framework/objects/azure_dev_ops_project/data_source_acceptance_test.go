package azure_dev_ops_project_test

import (
	"os"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudAzureDevOpsProject(t *testing.T) {
	//TODO: Remove both env var checks when this gets configurted in CI and the variables are parameterized
	if os.Getenv("CI") != "" {
		t.Skip("Skipping Azure DevOps Project datasource test in CI " +
			"until Azure integration and a personal access token are available")
	}

	if os.Getenv("DBT_CLOUD_PERSONAL_ACCESS_TOKEN") == "" {
		t.Skip("Skipping Azure DevOps Project datasource because no personal access token is available")
	}

	//TODO: Parameterize these values when a standard method is available for parameterization
	adoProjectName := "dbt-cloud-ado-project"
	personalAccessToken := os.Getenv("DBT_CLOUD_PERSONAL_ACCESS_TOKEN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest_helper.TestAccPreCheck(t) },

		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigVariables: config.Variables{
					"dbt_token":        config.StringVariable(personalAccessToken),
					"ado_project_name": config.StringVariable(adoProjectName),
				},
				Config: `
					variable "dbt_token" {
						type = string
						sensitive = true
					}

					provider "dbtcloud" {
						token = var.dbt_token
					}

					variable "ado_project_name" {
						type = string
					}
					
					data dbtcloud_azure_dev_ops_project test {
						name = var.ado_project_name
					}
				`,
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_azure_dev_ops_project.test",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_azure_dev_ops_project.test",
						"url",
					),
				),
			},
		},
	})

}
