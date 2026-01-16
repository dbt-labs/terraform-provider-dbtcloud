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

// TestAccDbtCloudJobResourceTimeoutSecondsBackwardCompatibility tests that the
// deprecated top-level timeout_seconds attribute still works for backward compatibility
func TestAccDbtCloudJobResourceTimeoutSecondsBackwardCompatibility(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				// Create job with deprecated timeout_seconds
				Config: testAccDbtCloudJobResourceTimeoutConfig(
					jobName,
					projectName,
					environmentName,
					"deprecated", // uses top-level timeout_seconds
					180,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "timeout_seconds", "180"),
				),
			},
			// Update the timeout value using deprecated attribute
			{
				Config: testAccDbtCloudJobResourceTimeoutConfig(
					jobName,
					projectName,
					environmentName,
					"deprecated",
					360,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "timeout_seconds", "360"),
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
				},
			},
		},
	})
}

// TestAccDbtCloudJobResourceExecutionTimeoutSeconds tests that the new
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
				// Create job with new execution.timeout_seconds
				Config: testAccDbtCloudJobResourceTimeoutConfig(
					jobName,
					projectName,
					environmentName,
					"execution", // uses execution block
					240,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execution.timeout_seconds", "240"),
					// When using execution block, deprecated timeout_seconds keeps its default (0)
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "timeout_seconds", "0"),
				),
			},
			// Update the timeout value using execution block
			{
				Config: testAccDbtCloudJobResourceTimeoutConfig(
					jobName,
					projectName,
					environmentName,
					"execution",
					480,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execution.timeout_seconds", "480"),
					// When using execution block, deprecated timeout_seconds keeps its default (0)
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "timeout_seconds", "0"),
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
					"execution",       // execution block is only set if user configured it
					"timeout_seconds", // after import, this gets API value since we don't know user's config preference
				},
			},
		},
	})
}

// TestAccDbtCloudJobResourceTimeoutSecondsPrecedence tests that when both
// deprecated timeout_seconds and execution.timeout_seconds are set,
// execution.timeout_seconds takes precedence
func TestAccDbtCloudJobResourceTimeoutSecondsPrecedence(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				// Create job with both timeout settings - execution should take precedence
				Config: testAccDbtCloudJobResourceTimeoutBothConfig(
					jobName,
					projectName,
					environmentName,
					100, // deprecated timeout_seconds (ignored when execution is set)
					300, // execution.timeout_seconds (should win)
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					// execution.timeout_seconds takes precedence and is used for the API call
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execution.timeout_seconds", "300"),
					// When execution block is used, timeout_seconds keeps its configured value
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "timeout_seconds", "100"),
				),
			},
		},
	})
}

// TestAccDbtCloudJobResourceTimeoutSecondsMigration tests migrating from
// deprecated timeout_seconds to the new execution.timeout_seconds
func TestAccDbtCloudJobResourceTimeoutSecondsMigration(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				// Start with deprecated timeout_seconds
				Config: testAccDbtCloudJobResourceTimeoutConfig(
					jobName,
					projectName,
					environmentName,
					"deprecated",
					180,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "timeout_seconds", "180"),
				),
			},
			// Migrate to execution.timeout_seconds
			{
				Config: testAccDbtCloudJobResourceTimeoutConfig(
					jobName,
					projectName,
					environmentName,
					"execution",
					180,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execution.timeout_seconds", "180"),
					// After migration, deprecated timeout_seconds goes back to default (0)
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "timeout_seconds", "0"),
				),
			},
		},
	})
}

// TestAccDbtCloudJobResourceTimeoutSecondsDefault tests that when neither
// timeout is specified, it defaults to 0
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
				Config: testAccDbtCloudJobResourceTimeoutConfig(
					jobName,
					projectName,
					environmentName,
					"none",
					0, // ignored when mode is "none"
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					// Default should be 0 (no timeout)
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "timeout_seconds", "0"),
				),
			},
		},
	})
}

func testAccDbtCloudJobResourceTimeoutConfig(
	jobName, projectName, environmentName, mode string,
	timeoutSeconds int,
) string {
	var timeoutConfig string

	switch mode {
	case "deprecated":
		// Use deprecated top-level timeout_seconds
		timeoutConfig = fmt.Sprintf("timeout_seconds = %d", timeoutSeconds)
	case "execution":
		// Use new execution attribute (SingleNestedAttribute uses = syntax, not block syntax)
		timeoutConfig = fmt.Sprintf(`
  execution = {
    timeout_seconds = %d
  }`, timeoutSeconds)
	case "none":
		// No timeout specified - use defaults
		timeoutConfig = ""
	default:
		panic("Invalid mode: " + mode)
	}

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
  %s
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, timeoutConfig)
}

func testAccDbtCloudJobResourceTimeoutBothConfig(
	jobName, projectName, environmentName string,
	deprecatedTimeout, executionTimeout int,
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
  timeout_seconds = %d
  execution = {
    timeout_seconds = %d
  }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, deprecatedTimeout, executionTimeout)
}
