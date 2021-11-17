package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudProjectDataSource(t *testing.T) {

	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := project(randomProjectName)

	check := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "project_id"),
		resource.TestCheckResourceAttr("data.dbt_cloud_project.test", "name", randomProjectName),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "connection_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "repository_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "state"),
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

func project(projectName string) string {
	return fmt.Sprintf(`
    resource "dbt_cloud_project" "test" {
		name = "%s"
		dbt_project_subdirectory = "/path"
	}

    data "dbt_cloud_project" "test" {
		project_id = dbt_cloud_project.test.id
	}
    `, projectName)
}
