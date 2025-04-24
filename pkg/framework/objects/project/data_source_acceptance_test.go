package project_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudProjectDataSource(t *testing.T) {
	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDataSourceConfig(randomProjectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dbtcloud_project.test", "id"),
					resource.TestCheckResourceAttr("data.dbtcloud_project.test", "name", randomProjectName),
					resource.TestCheckResourceAttrSet("data.dbtcloud_project.test", "state"),

					resource.TestCheckResourceAttrSet("data.dbtcloud_project.test_with_name", "id"),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_project.test_with_name",
						"name",
						randomProjectName,
					),
					resource.TestCheckResourceAttrSet("data.dbtcloud_project.test_with_name", "state"),
				),
			},
		},
	})
}

func testAccProjectDataSourceConfig(projectName string) string {
	return fmt.Sprintf(`
    resource "dbtcloud_project" "test" {
		name = "%s"
		dbt_project_subdirectory = "/path"
	}

    data "dbtcloud_project" "test" {
		id = dbtcloud_project.test.id
	}

	data "dbtcloud_project" "test_with_name" {
		name = dbtcloud_project.test.name
	}
    `, projectName)
}
