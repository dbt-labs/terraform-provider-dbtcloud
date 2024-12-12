package azure_dev_ops_repository_test

import (
	"os"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudAzureDevOpsRepository(t *testing.T) {
	//TODO: Remove both env var checks when this gets configurted in CI and the variables are parameterized
	if os.Getenv("CI") != "" {
		t.Skip("Skipping Azure DevOps Repository datasource test in CI " +
			"until Azure integration and a personal access token are available")
	}

	if os.Getenv("DBT_CLOUD_PERSONAL_ACCESS_TOKEN") == "" {
		t.Skip("Skipping Azure DevOps Repository datasource because no personal access token is available")
	}

	//TODO: Parameterize these values when a standard method is available for parameterization
	adoProjectName := "dbt-cloud-ado-project"
	adoRepoName := "dbt-cloud-ado-repo"
	personalAccessToken := os.Getenv("DBT_CLOUD_PERSONAL_ACCESS_TOKEN")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest_helper.TestAccPreCheck(t) },

		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigVariables: config.Variables{
					"dbt_token":           config.StringVariable(personalAccessToken),
					"ado_project_name":    config.StringVariable(adoProjectName),
					"ado_repository_name": config.StringVariable(adoRepoName),
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

					variable "ado_repository_name" {
						type = string
					}
					
					data dbtcloud_azure_dev_ops_project test {
						name = var.ado_project_name
					}

					data dbtcloud_azure_dev_ops_repository test {
						name = var.ado_repository_name
						azure_dev_ops_project_id = data.dbtcloud_azure_dev_ops_project.test.id
					}
				`,
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_azure_dev_ops_repository.test",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_azure_dev_ops_repository.test",
						"details_url",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_azure_dev_ops_repository.test",
						"remote_url",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_azure_dev_ops_repository.test",
						"web_url",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_azure_dev_ops_repository.test",
						"default_branch",
					),
				),
			},
		},
	})

}
