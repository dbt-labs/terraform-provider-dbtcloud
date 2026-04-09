package job_completion_trigger_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudJobCompletionTriggerResource(t *testing.T) {
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	jobNameA := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	jobNameB := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobCompletionTriggerResourceConfig(projectName, jobNameA, jobNameB),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_job_completion_trigger.test",
						"id",
					),
					resource.TestCheckResourceAttrPair(
						"dbtcloud_job_completion_trigger.test",
						"job_id",
						"dbtcloud_job.job_b",
						"id",
					),
					resource.TestCheckResourceAttrPair(
						"dbtcloud_job_completion_trigger.test",
						"trigger_job_id",
						"dbtcloud_job.job_a",
						"id",
					),
					resource.TestCheckResourceAttrPair(
						"dbtcloud_job_completion_trigger.test",
						"project_id",
						"dbtcloud_project.test_project",
						"id",
					),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_job_completion_trigger.test",
						"statuses.*",
						"success",
					),
				),
			},
			{
				ResourceName:            "dbtcloud_job_completion_trigger.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resource_metadata"},
			},
		},
	})
}

func testAccDbtCloudJobCompletionTriggerResourceConfig(projectName, jobNameA, jobNameB string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_environment" "test_env" {
  dbt_version    = "%s"
  name           = "test"
  project_id     = dbtcloud_project.test_project.id
  type           = "deployment"
}

resource "dbtcloud_job" "job_a" {
  name           = "%s"
  project_id     = dbtcloud_project.test_project.id
  environment_id = dbtcloud_environment.test_env.environment_id
  execute_steps  = ["dbt run"]
  triggers = {
    github_webhook      = false
    git_provider_webhook = false
    schedule            = false
  }
}

resource "dbtcloud_job" "job_b" {
  name           = "%s"
  project_id     = dbtcloud_project.test_project.id
  environment_id = dbtcloud_environment.test_env.environment_id
  execute_steps  = ["dbt run"]
  triggers = {
    github_webhook      = false
    git_provider_webhook = false
    schedule            = false
  }
}

resource "dbtcloud_job_completion_trigger" "test" {
  job_id         = dbtcloud_job.job_b.id
  trigger_job_id = dbtcloud_job.job_a.id
  project_id     = dbtcloud_project.test_project.id
  statuses       = ["success"]
}
`, projectName, acctest_config.DBT_CLOUD_VERSION, jobNameA, jobNameB)
}
