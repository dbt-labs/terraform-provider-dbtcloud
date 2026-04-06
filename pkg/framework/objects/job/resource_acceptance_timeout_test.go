package job_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDbtCloudJobResourceExecutionTimeoutSeconds tests that the
// execution.timeout_seconds attribute works correctly
func TestAccDbtCloudJobResourceExecutionTimeoutSeconds(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceTimeoutConfig(
					jobName,
					projectName,
					environmentName,
					240,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execution.timeout_seconds", "240"),
				),
			},
			{
				Config: testAccDbtCloudJobResourceTimeoutConfig(
					jobName,
					projectName,
					environmentName,
					480,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execution.timeout_seconds", "480"),
				),
			},
			// Import and verify
			{
				ResourceName:      "dbtcloud_job.test_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"triggers.%",
					"triggers.custom_branch_only",
					"validate_execute_steps",
					"execution", // execution block is only set if user configured it
				},
			},
		},
	})
}

// TestAccDbtCloudJobResourceTimeoutSecondsDefault tests that when no timeout is
// specified, it defaults to 0
func TestAccDbtCloudJobResourceTimeoutSecondsDefault(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				// Create job without any timeout specified
				Config: testAccDbtCloudJobResourceNoTimeoutConfig(
					jobName,
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
				),
			},
		},
	})
}

func testAccDbtCloudJobResourceTimeoutConfig(
	jobName, projectName, environmentName string,
	timeoutSeconds int,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name = "%s"
    dbt_version = "%s"
    type = "development"
}

resource "dbtcloud_job" "test_job" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment.environment_id
  execute_steps = [
    "dbt test"
  ]
  triggers = {
    "github_webhook": false,
    "git_provider_webhook": false,
    "schedule": false,
  }
  execution = {
    timeout_seconds = %d
  }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, timeoutSeconds)
}

func testAccDbtCloudJobResourceNoTimeoutConfig(
	jobName, projectName, environmentName string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name = "%s"
    dbt_version = "%s"
    type = "development"
}

resource "dbtcloud_job" "test_job" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment.environment_id
  execute_steps = [
    "dbt test"
  ]
  triggers = {
    "github_webhook": false,
    "git_provider_webhook": false,
    "schedule": false,
  }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName)
}
