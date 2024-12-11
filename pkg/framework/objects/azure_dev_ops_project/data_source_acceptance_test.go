package azure_dev_ops_project_test

import (
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudAzureDevOpsProject(t *testing.T) {
	if acctest_helper.IsDbtCloudPR() {
		t.Skip("Skipping Azure DevOps Project datasource test in CI until a Personal Access Token is available")
	}

	adoProjectName := "dbt-cloud-ado-project"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest_helper.TestAccPreCheck(t) },

		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigVariables: config.Variables{
					"dbt_token":        config.StringVariable("this-will-come-from-somewhere-else-eventually"), // Cannot be a Service Token
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
