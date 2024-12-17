package job_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDbtCloudJobDataSource(t *testing.T) {

	randomJobName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := jobs(randomJobName)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.dbtcloud_job.test", "job_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_job.test", "project_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_job.test", "environment_id"),
		resource.TestCheckResourceAttr("data.dbtcloud_job.test", "name", randomJobName),
		resource.TestCheckResourceAttr("data.dbtcloud_job.test", "timeout_seconds", "180"),
		resource.TestCheckResourceAttr("data.dbtcloud_job.test", "triggers_on_draft_pr", "false"),
		resource.TestCheckResourceAttr(
			"data.dbtcloud_job.test",
			"job_completion_trigger_condition.#",
			"0",
		),
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

func jobs(jobName string) string {
	return fmt.Sprintf(`
    resource "dbtcloud_project" "test_project" {
        name = "jobs_test_project"
    }

    resource "dbtcloud_environment" "test_environment" {
        project_id = dbtcloud_project.test_project.id
        name = "job_test_env"
        dbt_version = "%s"
        type = "development"
    }

    resource "dbtcloud_job" "test_job" {
        name = "%s"
        project_id = dbtcloud_project.test_project.id
        environment_id = dbtcloud_environment.test_environment.environment_id
        execute_steps = [
            "dbt run"
        ]
        triggers = {
          "github_webhook" : false,
          "schedule" : false,
          "git_provider_webhook": false
        }
        timeout_seconds = 180
    }

    data "dbtcloud_job" "test" {
        job_id = dbtcloud_job.test_job.id
        project_id = dbtcloud_project.test_project.id
    }
    `, acctest_helper.DBT_CLOUD_VERSION, jobName)
}
