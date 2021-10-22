package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudProjectDataSource(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := fmt.Sprintf(`
			data "dbt_cloud_project" "test" {
				project_id = "%s"
			}
		`, randomID)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbt_cloud_job.test", "project_id", randomID),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "name"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "id"),
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
