package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudJobDataSource(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := fmt.Sprintf(`
			data "dbt_cloud_job" "test" {
				job_id = "%s"
			}
		`, randomID)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbt_cloud_job.test", "job_id", randomID),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_job.test", "project_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_job.test", "environment_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_job.test", "name"),
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
