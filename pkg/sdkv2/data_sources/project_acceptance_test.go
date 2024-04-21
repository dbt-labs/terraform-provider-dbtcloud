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
		resource.TestCheckResourceAttrSet("data.dbtcloud_project.test", "project_id"),
		resource.TestCheckResourceAttr("data.dbtcloud_project.test", "name", randomProjectName),
		resource.TestCheckResourceAttrSet("data.dbtcloud_project.test", "connection_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_project.test", "repository_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_project.test", "state"),

		resource.TestCheckResourceAttrSet("data.dbtcloud_project.test_with_name", "project_id"),
		resource.TestCheckResourceAttr("data.dbtcloud_project.test_with_name", "name", randomProjectName),
		resource.TestCheckResourceAttrSet("data.dbtcloud_project.test_with_name", "connection_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_project.test_with_name", "repository_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_project.test_with_name", "state"),
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
    resource "dbtcloud_project" "test" {
		name = "%s"
		dbt_project_subdirectory = "/path"
	}

    data "dbtcloud_project" "test" {
		project_id = dbtcloud_project.test.id
	}

	data "dbtcloud_project" "test_with_name" {
		name = dbtcloud_project.test.name
	}
    `, projectName)
}
