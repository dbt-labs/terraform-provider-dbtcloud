package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudEnvironmentDataSource(t *testing.T) {

	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	randomEnvironmentName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := environment(randomProjectName, randomEnvironmentName)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbtcloud_environment.test", "name", randomEnvironmentName),
		resource.TestCheckResourceAttrSet("data.dbtcloud_environment.test", "environment_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_environment.test", "project_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_environment.test", "is_active"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_environment.test", "credential_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_environment.test", "dbt_version"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_environment.test", "type"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_environment.test", "use_custom_branch"),
		resource.TestCheckResourceAttr("data.dbtcloud_environment.test", "custom_branch", "customBranchName"),
	)

	resource.ParallelTest(t, resource.TestCase{
		Providers: providers(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  check,
			},
		},
	})
}

func environment(projectName, environmentName string) string {
	return fmt.Sprintf(`
    resource "dbtcloud_project" "test_project" {
        name = "%s"
    }

    resource "dbtcloud_environment" "test_environment" {
        project_id = dbtcloud_project.test_project.id
        name = "%s"
        dbt_version = "%s"
        type = "development"
        use_custom_branch = true
        custom_branch = "customBranchName"
    }

    data "dbtcloud_environment" "test" {
        project_id = dbtcloud_project.test_project.id
        environment_id = dbtcloud_environment.test_environment.environment_id
    }
    `, projectName, environmentName, DBT_CLOUD_VERSION)
}
