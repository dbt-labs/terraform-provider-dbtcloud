package project_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDbtCloudJobsDataSource(t *testing.T) {

	projectName := acctest.RandStringFromCharSet(19, acctest.CharSetAlphaNum)
	projectName1 := fmt.Sprintf("%s1", projectName)
	projectName2 := fmt.Sprintf("%s2", projectName)

	config := jobs(projectName, projectName1, projectName2)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbtcloud_projects.test", "projects.#", "2"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_projects.test", "projects.0.id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_projects.test", "projects.0.name"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_projects.test", "projects.1.id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_projects.test", "projects.1.name"),
	)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  check,
			},
		},
	})
}

func jobs(projectName string, projectName1 string, projectName2 string) string {
	return fmt.Sprintf(`
    resource "dbtcloud_project" "test_project1" {
        name = "%s"
    }

	resource "dbtcloud_project" "test_project2" {
        name = "%s"
    }

	data dbtcloud_projects test {
		name_contains = "%s"

		depends_on = [
			dbtcloud_project.test_project1,
			dbtcloud_project.test_project2,
		]
	}

    `, projectName1, projectName2, projectName)
}
