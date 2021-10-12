package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudJobResource(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := fmt.Sprintf(`
			resource "dbt_cloud_job" "test" {
				name = "dbt-cloud-job-%s"
				project_id = 123
				environment_id = 789
				execute_steps = [
				    "dbt run",
				    "dbt test"
				]
				dbt_version = "0.20.0"
				is_active = true
				num_threads = 5
				target_name = "target"
				generate_docs = true
				run_generate_sources = true
				triggers = {
				    "github_webhook": true,
				    "schedule": true,
				    "custom_branch_only": true
				}
			}
		`, randomID)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("dbt_cloud_job.test", "job_id"),
		resource.TestCheckResourceAttr("dbt_cloud_job.test", "project_id", "123"),
		resource.TestCheckResourceAttr("dbt_cloud_job.test", "environment_id", "789"),
		resource.TestCheckResourceAttr("dbt_cloud_job.test", "name", fmt.Sprintf("dbt-cloud-job-%s", randomID)),
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
