package data_sources_test


import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDbtCloudJobDataSource(t *testing.T) {

	randomJobName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := jobs(randomJobName)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.dbt_cloud_job.test", "job_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_job.test", "project_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_job.test", "environment_id"),
		resource.TestCheckResourceAttr("data.dbt_cloud_job.test", "name", randomJobName),
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

func jobs(jobName string) string {
    return fmt.Sprintf(`
    resource "dbt_cloud_project" "test_project" {
        name = "jobs_test_project"
    }

    resource "dbt_cloud_environment" "test_environment" {
        project_id = dbt_cloud_project.test_project.id
        name = "job_test_env"
        dbt_version = "0.21.0"
        type = "development"
    }

    resource "dbt_cloud_job" "test_job" {
        name = "%s"
        project_id = dbt_cloud_project.test_project.id
        environment_id = dbt_cloud_environment.test_environment.environment_id
        execute_steps = [
            "dbt run"
        ]
        triggers = {
          "custom_branch_only" : false,
          "github_webhook" : false,
          "schedule" : false,
          "git_provider_webhook": false
        }
    }

    data "dbt_cloud_job" "test" {
        job_id = dbt_cloud_job.test_job.id
        project_id = dbt_cloud_project.test_project.id
    }
    `, jobName)
}
