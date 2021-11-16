package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDbtCloudProjectResource(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := fmt.Sprintf(`
			resource "dbt_cloud_project" "test" {
				name = "dbt-cloud-project-%s"
				dbt_project_subdirectory = "/this-way/for/DBT"
			}
		`, randomID)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("dbt_cloud_project.test", "ID"),
		resource.TestCheckResourceAttr("dbt_cloud_project.test", "name", fmt.Sprintf("dbt-cloud-project-%s", randomID)),
		resource.TestCheckResourceAttr("dbt_cloud_project.test", "dbt_project_subdirectory", "/this-way/for/DBT"),
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
