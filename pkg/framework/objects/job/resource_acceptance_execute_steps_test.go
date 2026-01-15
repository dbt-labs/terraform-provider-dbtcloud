package job_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudJobResourceExecuteStepsValidation(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Test: Invalid command should fail
			{
				Config: testAccDbtCloudJobResourceExecuteStepsConfig(
					jobName,
					projectName,
					environmentName,
					[]string{"invalid command"},
				),
				ExpectError: regexp.MustCompile(`invalid command`),
			},
		},
	})
}

func TestAccDbtCloudJobResourceExecuteStepsDuplicateFlag(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Test: Duplicate flag should fail
			{
				Config: testAccDbtCloudJobResourceExecuteStepsConfig(
					jobName,
					projectName,
					environmentName,
					[]string{"dbt --warn-error --warn-error run"},
				),
				ExpectError: regexp.MustCompile(`flag .* can only be used once`),
			},
		},
	})
}

func TestAccDbtCloudJobResourceExecuteStepsValid(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Test: Valid commands should succeed
			{
				Config: testAccDbtCloudJobResourceExecuteStepsConfig(
					jobName,
					projectName,
					environmentName,
					[]string{"dbt run", "dbt test"},
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execute_steps.#", "2"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execute_steps.0", "dbt run"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execute_steps.1", "dbt test"),
				),
			},
			// Update: Change to different valid commands with flags
			{
				Config: testAccDbtCloudJobResourceExecuteStepsConfig(
					jobName,
					projectName,
					environmentName,
					[]string{"dbt --warn-error build", "dbt docs generate"},
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execute_steps.#", "2"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execute_steps.0", "dbt --warn-error build"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execute_steps.1", "dbt docs generate"),
				),
			},
			// IMPORT
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

func TestAccDbtCloudJobResourceExecuteStepsMultipleFlags(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Test: Multiple different flags should succeed
			{
				Config: testAccDbtCloudJobResourceExecuteStepsConfig(
					jobName,
					projectName,
					environmentName,
					[]string{"dbt --warn-error --fail-fast run"},
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "execute_steps.0", "dbt --warn-error --fail-fast run"),
				),
			},
			// IMPORT
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

func TestAccDbtCloudJobResourceExecuteStepsMultiLine(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	// Multi-line command as shown in issue #600
	multiLineCommand := `dbt run --select
		my_project,config.materialized:view,state:modified,tag:my_tag
		my_project,config.materialized:table,tag:my_tag
		my_project,config.materialized:incremental,tag:my_tag`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Test: Multi-line commands should succeed (issue #600)
			{
				Config: testAccDbtCloudJobResourceExecuteStepsMultiLineConfig(
					jobName,
					projectName,
					environmentName,
					multiLineCommand,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
				),
			},
			// IMPORT
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

func testAccDbtCloudJobResourceExecuteStepsConfig(
	jobName, projectName, environmentName string,
	executeSteps []string,
) string {
	// Build execute_steps array string
	stepsStr := "["
	for i, step := range executeSteps {
		if i > 0 {
			stepsStr += ","
		}
		stepsStr += `"` + step + `"`
	}
	stepsStr += "]"

	return `
resource "dbtcloud_project" "test_job_project" {
    name = "` + projectName + `"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name = "` + environmentName + `"
    dbt_version = "` + acctest_config.DBT_CLOUD_VERSION + `"
    type = "deployment"
}

resource "dbtcloud_job" "test_job" {
    name = "` + jobName + `"
    project_id = dbtcloud_project.test_job_project.id
    environment_id = dbtcloud_environment.test_job_environment.environment_id
    execute_steps = ` + stepsStr + `
    validate_execute_steps = true
    triggers = {
        "github_webhook": false,
        "git_provider_webhook": false,
        "schedule": false
    }
}
`
}

func testAccDbtCloudJobResourceExecuteStepsMultiLineConfig(
	jobName, projectName, environmentName string,
	multiLineStep string,
) string {
	return `
resource "dbtcloud_project" "test_job_project" {
    name = "` + projectName + `"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name = "` + environmentName + `"
    dbt_version = "` + acctest_config.DBT_CLOUD_VERSION + `"
    type = "deployment"
}

resource "dbtcloud_job" "test_job" {
    name = "` + jobName + `"
    project_id = dbtcloud_project.test_job_project.id
    environment_id = dbtcloud_environment.test_job_environment.environment_id
    execute_steps = [<<-EOT
` + multiLineStep + `
EOT
    ]
    validate_execute_steps = true
    triggers = {
        "github_webhook": false,
        "git_provider_webhook": false,
        "schedule": false
    }
}
`
}
