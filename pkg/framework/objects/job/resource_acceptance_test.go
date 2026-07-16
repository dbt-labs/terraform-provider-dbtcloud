package job_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudJobResource(t *testing.T) {

	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	jobName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	// for deferral
	jobName3 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	// for job chaining
	jobName4 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	var configDeferral string
	var checkDeferral resource.TestCheckFunc

	configDeferral = testAccDbtCloudJobResourceDeferringConfig(
		jobName,
		jobName2,
		jobName3,
		projectName,
		environmentName,
		"env",
	)
	checkDeferral = resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
		testAccCheckDbtCloudJobExists("dbtcloud_job.test_job_2"),
		testAccCheckDbtCloudJobExists("dbtcloud_job.test_job_3"),
		resource.TestCheckResourceAttrSet("dbtcloud_job.test_job_2", "deferring_environment_id"),
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceBasicConfig(
					jobName,
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudJobResourceBasicConfig(
					jobName2,
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName2),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudJobResourceFullConfig(
					jobName2,
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName2),
					resource.TestCheckResourceAttr(
						"dbtcloud_job.test_job",
						"dbt_version",
						acctest_config.DBT_CLOUD_VERSION,
					),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "target_name", "test"),
					resource.TestCheckResourceAttr(
						"dbtcloud_job.test_job",
						"timeout_seconds",
						"180",
					),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "project_id"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "environment_id"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "is_active"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "num_threads"),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_job.test_job",
						"run_generate_sources",
					),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "generate_docs"),
				),
			},
			// JOB CHAINING
			{
				Config: testAccDbtCloudJobResourceJobChaining(
					jobName2,
					projectName,
					environmentName,
					jobName4,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job_4"),
					resource.TestCheckResourceAttr(
						"dbtcloud_job.test_job_4",
						"job_completion_trigger_condition.#",
						"1",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_job.test_job_4",
						"job_completion_trigger_condition.0.job_id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_job.test_job_4",
						"job_completion_trigger_condition.0.project_id",
					),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_job.test_job_4",
						"job_completion_trigger_condition.0.statuses.*",
						"error",
					),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_job.test_job_4",
						"job_completion_trigger_condition.0.statuses.*",
						"success",
					),
				),
			},
			// DEFERRING JOBS (depends on whether DBT_LEGACY_JOB_DEFERRAL is set, e.g. whether the new CI is set)
			{
				Config: configDeferral,
				Check:  checkDeferral,
			},
			// REMOVE DEFERRAL
			{
				Config: testAccDbtCloudJobResourceFullConfig(
					jobName2,
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName2),
					resource.TestCheckResourceAttr(
						"dbtcloud_job.test_job",
						"dbt_version",
						acctest_config.DBT_CLOUD_VERSION,
					),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "target_name", "test"),
					resource.TestCheckResourceAttr(
						"dbtcloud_job.test_job",
						"timeout_seconds",
						"180",
					),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "project_id"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "environment_id"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "is_active"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "num_threads"),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_job.test_job",
						"run_generate_sources",
					),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "generate_docs"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.test_job",
				ImportState:       true,
				ImportStateVerify: true,
				// we don't check triggers.custom_branch_only as we currently allow people to keep triggers.custom_branch_only in their config to not break peopple's Terraform project
				ImportStateVerifyIgnore: []string{
					"triggers.%",
					"triggers.custom_branch_only",
					"validate_execute_steps",
				},
			},
		},
	})
}

func TestAccDbtCloudJobResourceTriggers(t *testing.T) {

	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceBasicConfigTriggers(
					jobName,
					projectName,
					environmentName,
					"git",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
				),
			},
			// MODIFY TRIGGERS
			{
				Config: testAccDbtCloudJobResourceBasicConfigTriggers(
					jobName,
					projectName,
					environmentName,
					"on_merge",
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
					"validate_execute_steps",
				},
			},
		},
	})
}

func TestAccDbtCloudJobCISettings(t *testing.T) {

	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceCISettings(
					jobName,
					projectName,
					environmentName,
					false,
					true,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "run_lint", "false"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "errors_on_lint_failure", "true"),
				),
			},
			// MODIFY LINTING
			{
				Config: testAccDbtCloudJobResourceCISettings(
					jobName,
					projectName,
					environmentName,
					true,
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "run_lint", "true"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "errors_on_lint_failure", "false"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.ci_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
				},
			},
		},
	})
}

func testAccDbtCloudJobResourceCISettings(
	jobName string,
	projectName string,
	environmentName string,
	runLint bool,
	errorOnLinFailure bool,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "project" {
    name = "%s"
}

resource "dbtcloud_environment" "ci_env" {
    project_id = dbtcloud_project.project.id
    name = "%s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "ci_job" {
    project_id = dbtcloud_project.project.id
    environment_id = dbtcloud_environment.ci_env.environment_id
    name = "%s"
    dbt_version = "%s"
	execute_steps = [
	"dbt build -s state:modified+ --fail-fast"
	]
    run_lint = %t
    errors_on_lint_failure = %t

	triggers = {
		"github_webhook" : true
		"git_provider_webhook" : true
        "schedule" : false
	}
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, acctest_config.DBT_CLOUD_VERSION, runLint, errorOnLinFailure)
}

func testAccDbtCloudJobResourceBasicConfig(jobName, projectName, environmentName string) string {
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

func testAccDbtCloudJobResourceFullConfig(jobName, projectName, environmentName string) string {
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

resource "dbtcloud_environment" "test_job_environment_new" {
    project_id = dbtcloud_project.test_job_project.id
    name = "DEPL %s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "test_job" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment_new.environment_id
  dbt_version = "%s"
  execute_steps = [
    "dbt test"
  ]
  triggers = {
    "github_webhook": false,
    "git_provider_webhook": false,
    "schedule": true,
    "custom_branch_only": false,
  }
  is_active = true
  num_threads = 37
  target_name = "test"
  run_generate_sources = true
  generate_docs = true
  schedule_type = "every_day"
  schedule_hours = [9, 17]
  timeout_seconds = 180
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, acctest_config.DBT_CLOUD_VERSION)
}

func testAccDbtCloudJobResourceJobChaining(
	jobName, projectName, environmentName, jobName4 string,
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

resource "dbtcloud_environment" "test_job_environment_new" {
    project_id = dbtcloud_project.test_job_project.id
    name = "DEPL %s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "test_job" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment_new.environment_id
  dbt_version = "%s"
  execute_steps = [
    "dbt test"
  ]
  triggers = {
    "github_webhook": false,
    "git_provider_webhook": false,
    "schedule": true,
    "custom_branch_only": true,
  }
  is_active = true
  num_threads = 37
  target_name = "test"
  run_generate_sources = true
  generate_docs = true
  schedule_type = "every_day"
  schedule_hours = [9, 17]
  timeout_seconds = 180
}

resource "dbtcloud_job" "test_job_4" {
	name        = "%s"
	project_id = dbtcloud_project.test_job_project.id
	environment_id = dbtcloud_environment.test_job_environment.environment_id
	execute_steps = [
	  "dbt build +my_model"
	]
	triggers = {
	  "github_webhook": false,
	  "git_provider_webhook": false,
	  "schedule": false,
	}
	job_completion_trigger_condition {
		job_id = dbtcloud_job.test_job.id
		project_id = dbtcloud_project.test_job_project.id
		statuses = ["error", "success"]
	}
  }
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, acctest_config.DBT_CLOUD_VERSION, jobName4)
}

func testAccDbtCloudJobResourceDeferringConfig(
	jobName, jobName2, jobName3, projectName, environmentName string,
	deferring string,
) string {
	deferParam := ""
	selfDefer := ""
	if deferring == "job" {
		deferParam = "deferring_job_id = dbtcloud_job.test_job.id"
		selfDefer = "self_deferring = true"
	} else if deferring == "env" {
		deferParam = "deferring_environment_id = dbtcloud_environment.test_job_environment_new.environment_id"
	}
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment_new" {
    project_id = dbtcloud_project.test_job_project.id
    name = "DEPL %s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "test_job" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment_new.environment_id
  dbt_version = "%s"
  execute_steps = [
    "dbt test"
  ]
  triggers = {
    "github_webhook": false,
    "git_provider_webhook": false,
    "schedule": true,
  }
  is_active = true
  num_threads = 37
  target_name = "test"
  run_generate_sources = true
  generate_docs = true
  schedule_type = "every_day"
  schedule_hours = [9, 17]
  triggers_on_draft_pr = true
}

resource "dbtcloud_job" "test_job_2" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment_new.environment_id
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

resource "dbtcloud_job" "test_job_3" {
	name        = "%s"
	project_id = dbtcloud_project.test_job_project.id
	environment_id = dbtcloud_environment.test_job_environment_new.environment_id
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
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, acctest_config.DBT_CLOUD_VERSION, jobName2, deferParam, jobName3, selfDefer)
}

func TestAccDbtCloudJobResourceSchedules(t *testing.T) {

	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceScheduleConfig(
					jobName,
					projectName,
					environmentName,
					"every_day",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
				),
			},
			// MODIFY SCHEDULE
			{
				Config: testAccDbtCloudJobResourceScheduleConfig(
					jobName,
					projectName,
					environmentName,
					"days_of_week",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
				),
			},
			// MODIFY SCHEDULE
			{
				Config: testAccDbtCloudJobResourceScheduleConfig(
					jobName,
					projectName,
					environmentName,
					"custom_cron",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
				),
			},
			// MODIFY SCHEDULE
			{
				Config: testAccDbtCloudJobResourceScheduleConfig(
					jobName,
					projectName,
					environmentName,
					"interval_cron",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "schedule_type", "interval_cron"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "schedule_interval", "5"),
				),
			},

			// IMPORT
			{
				ResourceName:      "dbtcloud_job.test_job",
				ImportState:       true,
				ImportStateVerify: true,
				// we don't check triggers.custom_branch_only as we currently allow people to keep triggers.custom_branch_only in their config to not break peopple's Terraform project
				ImportStateVerifyIgnore: []string{
					"triggers.%",
					"triggers.custom_branch_only",
					"validate_execute_steps",
				},
			},
		},
	})
}

func testAccDbtCloudJobResourceScheduleConfig(
	jobName, projectName, environmentName, scheduleType string,
) string {

	scheduleConfig := ""
	if scheduleType == "every_day" {
		scheduleConfig = `
		schedule_type = "every_day"
		schedule_hours = [1,2,3]`
	} else if scheduleType == "days_of_week" {
		scheduleConfig = `
		schedule_type = "days_of_week"
		schedule_interval = 2
  		schedule_days = [1,4]`
	} else if scheduleType == "custom_cron" {
		scheduleConfig = `	
		schedule_cron = "0 21 * * *"
		schedule_type = "custom_cron"`
	} else if scheduleType == "interval_cron" {
		scheduleConfig = `
		schedule_type = "interval_cron"
		schedule_days = [0,1,2,3,4,5,6]
		schedule_interval = 5`
	} else {
		panic("Incorrect schedule type")
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
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, scheduleConfig)
}

func testAccDbtCloudJobResourceBasicConfigTriggers(
	jobName, projectName, environmentName, trigger string,
) string {

	git_trigger := "false"
	schedule_trigger := "false"
	on_merge_trigger := "false"
	run_compare_changes := "false"
	deferringConfig := ""

	if trigger == "git" {
		git_trigger = "true"
		deferringConfig = "deferring_environment_id = dbtcloud_environment.test_job_environment.environment_id"
		if !acctest_config.IsDbtCloudPR() {
			// we don't want to activate it in Cloud PRs as the setting need to be ON
			// TODO: When TF supports account settings, activate the setting in this test and remove this logic
			run_compare_changes = "true"
		}
	}
	if trigger == "schedule" {
		schedule_trigger = "true"
	}
	if trigger == "on_merge" {
		on_merge_trigger = "true"
	}

	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name = "%s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "test_job" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment.environment_id
  execute_steps = [
    "dbt run"
  ]
  triggers = {
    "github_webhook": %s,
    "git_provider_webhook": %s,
    "schedule": %s,
	"on_merge": %s
  }
  run_compare_changes = %s
  %s
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, git_trigger, git_trigger, schedule_trigger, on_merge_trigger, run_compare_changes, deferringConfig)
}

func testAccCheckDbtCloudJobExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetJob(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudJobHasDbtState(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("issue getting the client: %w", err)
		}
		remoteJob, err := apiClient.GetJob(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching job %s: %w", rs.Primary.ID, err)
		}
		if len(remoteJob.CostOptimizationFeatures) != 1 || remoteJob.CostOptimizationFeatures[0] != "dbt_state" {
			return fmt.Errorf("remote job cost_optimization_features = %#v, want [\"dbt_state\"]", remoteJob.CostOptimizationFeatures)
		}
		return nil
	}
}

func testAccCheckDbtCloudJobHasNoCostOptimizationFeatures(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("issue getting the client: %w", err)
		}
		remoteJob, err := apiClient.GetJob(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching job %s: %w", rs.Primary.ID, err)
		}
		if len(remoteJob.CostOptimizationFeatures) != 0 {
			return fmt.Errorf("remote job cost_optimization_features = %#v, want []", remoteJob.CostOptimizationFeatures)
		}
		return nil
	}
}

func skipDbtStateAcceptanceTest(t *testing.T) {
	t.Helper()
	// The shared CI acceptance account is not entitled to dbt State. Run these
	// tests locally only with an account that has the dbt State feature enabled.
	if acctest_config.IsCI() {
		t.Skip("Skipping in CI: requires an account with dbt State enabled.")
	}
}

func testAccCheckDbtCloudJobDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_job" {
			continue
		}
		_, err := apiClient.GetJob(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Job still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}

func TestAccDbtCloudJobResourceJobTypeAndCompareChanges(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceJobTypeAndCompareChangesConfig(
					jobName,
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "job_type", "ci"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "compare_changes_flags", "--select state:modified+"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "triggers.github_webhook", "false"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "triggers.git_provider_webhook", "false"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "triggers.schedule", "false"),
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

func testAccDbtCloudJobResourceJobTypeAndCompareChangesConfig(jobName, projectName, environmentName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name = "%s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "test_job" {
    name = "%s"
    project_id = dbtcloud_project.test_job_project.id
    environment_id = dbtcloud_environment.test_job_environment.environment_id
	deferring_environment_id = dbtcloud_environment.test_job_environment.environment_id
    execute_steps = [
        "dbt build"
    ]
    triggers = {
        "github_webhook": false,
        "git_provider_webhook": false,
        "schedule": false
    }
    job_type = "ci"
    run_compare_changes = true
    compare_changes_flags = "--select state:modified+"
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName)
}

func TestAccDbtCloudJobResourceIntervalCron(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceIntervalCronConfig(
					jobName,
					projectName,
					environmentName,
					5,
					[]int{0, 1, 2, 3, 4, 5, 6},
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "schedule_type", "interval_cron"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "schedule_interval", "5"),
				),
			},
			// MODIFY INTERVAL
			{
				Config: testAccDbtCloudJobResourceIntervalCronConfig(
					jobName,
					projectName,
					environmentName,
					10,
					[]int{0, 1, 2, 3, 4, 5, 6},
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "schedule_type", "interval_cron"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "schedule_interval", "10"),
				),
			},
			// MODIFY DAYS
			{
				Config: testAccDbtCloudJobResourceIntervalCronConfig(
					jobName,
					projectName,
					environmentName,
					10,
					[]int{1, 3, 5},
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "schedule_type", "interval_cron"),
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

// TestAccDbtCloudJobCostOptimizationFeatures tests creating a job with
// cost_optimization_features set (the preferred replacement for force_node_selection).
//
// Account requirements: the test account must either (a) have account-level SAO
// enforcement enabled so force_node_selection is forced to false, or (b) be running
// against an environment whose dbt_version is in FUSION_VERSIONS (currently only
// "latest-fusion") AND have State-Aware Orchestration available. On accounts that
// satisfy neither condition the dbt Cloud API silently rewrites force_node_selection
// back to true, which makes Terraform see an inconsistent result after apply
// (plan: ["state_aware_orchestration"], actual: []).
func TestAccDbtCloudJobCostOptimizationFeatures(t *testing.T) {
	// SAO requires an account entitled to State-Aware Orchestration (the
	// orc-2664-sao-beta flag) running on a Fusion dbt_version in a
	// staging/production environment. The CI account does not have it, so the API
	// silently rewrites the value, producing an "inconsistent result after apply".
	// Skip in CI; run locally against an SAO-enabled account.
	if acctest_config.IsCI() {
		t.Skip("Skipping in CI: requires an account with State-Aware Orchestration enabled.")
	}

	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	jobNameUpdated := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// 1. Create job with cost_optimization_features = ["state_aware_orchestration"]
			{
				Config: testAccDbtCloudJobCostOptimizationFeaturesConfig(
					jobName, projectName, environmentName, `["state_aware_orchestration"]`,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_job.test_job",
						"cost_optimization_features.*",
						"state_aware_orchestration",
					),
					// force_node_selection is not set in config; the API derives it from
					// cost_optimization_features (SAO present => force_node_selection = false).
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "force_node_selection", "false"),
				),
			},
			// 2. Update job name while keeping cost_optimization_features stable
			{
				Config: testAccDbtCloudJobCostOptimizationFeaturesConfig(
					jobNameUpdated, projectName, environmentName, `["state_aware_orchestration"]`,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_job.test_job",
						"cost_optimization_features.*",
						"state_aware_orchestration",
					),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "force_node_selection", "false"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobNameUpdated),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.test_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
				},
			},
		},
	})
}

// TestAccDbtCloudJobDbtStateCostOptimizationFeature tests creating a job with
// cost_optimization_features = ["dbt_state"], the migration path from Select All
// Optimizations (SAO) to dbt State.
//
// Account requirements: the test account must have dbt State enabled (the
// ORC_3638_ENABLE_DBT_STATE feature flag). Unlike SAO, dbt_state is
// environment-independent and does not require a Fusion dbt_version or a
// staging/production environment; it is also supported on CI and Merge jobs.
//
// The dbt platform API gives dbt_state precedence and collapses the set to
// ["dbt_state"], so the provider rejects mixing it with other features at plan
// time. It shares the dbt State acceptance-test entitlement gate, so it only
// runs against an account where dbt State is available.
func TestAccDbtCloudJobDbtStateCostOptimizationFeature(t *testing.T) {
	skipDbtStateAcceptanceTest(t)

	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	jobNameUpdated := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// 1. Create job with cost_optimization_features = ["dbt_state"]
			{
				Config: testAccDbtCloudJobCostOptimizationFeaturesConfig(
					jobName, projectName, environmentName, `["dbt_state"]`,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_job.test_job",
						"cost_optimization_features.*",
						"dbt_state",
					),
					// dbt_state implies SAO is on (force_node_selection = false)
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "force_node_selection", "false"),
				),
			},
			// 2. Update job name while keeping cost_optimization_features stable
			{
				Config: testAccDbtCloudJobCostOptimizationFeaturesConfig(
					jobNameUpdated, projectName, environmentName, `["dbt_state"]`,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_job.test_job",
						"cost_optimization_features.*",
						"dbt_state",
					),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobNameUpdated),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.test_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
				},
			},
		},
	})
}

// TestAccDbtCloudJobCIWithDbtStateCostOptimizationFeature reproduces #715:
// a github_webhook-triggered CI job must persist dbt_state through Create,
// unrelated Update, and an Import/Read refresh. The remote checks prove the
// provider sent dbt_state rather than merely retaining Terraform's planned set.
func TestAccDbtCloudJobCIWithDbtStateCostOptimizationFeature(t *testing.T) {
	skipDbtStateAcceptanceTest(t)

	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	jobNameUpdated := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobCIWithDbtStateConfig(
					jobName, projectName, environmentName, `["dbt_state"]`,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "job_type", "ci"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "cost_optimization_features.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_job.ci_job",
						"cost_optimization_features.*",
						"dbt_state",
					),
					testAccCheckDbtCloudJobHasDbtState("dbtcloud_job.ci_job"),
				),
			},
			// Rename only: dbt_state must remain on the remote job during an unrelated update.
			{
				Config: testAccDbtCloudJobCIWithDbtStateConfig(
					jobNameUpdated, projectName, environmentName, `["dbt_state"]`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "name", jobNameUpdated),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "cost_optimization_features.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_job.ci_job",
						"cost_optimization_features.*",
						"dbt_state",
					),
					testAccCheckDbtCloudJobHasDbtState("dbtcloud_job.ci_job"),
				),
			},
			// IMPORT / REFRESH: ImportStateVerify invokes Read with no prior feature state.
			{
				ResourceName:      "dbtcloud_job.ci_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "cost_optimization_features.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_job.ci_job",
						"cost_optimization_features.*",
						"dbt_state",
					),
					testAccCheckDbtCloudJobHasDbtState("dbtcloud_job.ci_job"),
				),
			},
			// Explicit clear: CI supports [] as a first-class dbt_state update.
			{
				Config: testAccDbtCloudJobCIWithDbtStateConfig(
					jobNameUpdated, projectName, environmentName, `[]`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "cost_optimization_features.#", "0"),
					testAccCheckDbtCloudJobHasNoCostOptimizationFeatures("dbtcloud_job.ci_job"),
				),
			},
		},
	})
}

// TestAccDbtCloudJobCIWithDeferringEnvironment tests that a CI job correctly
// preserves deferring_environment_id after apply (Change 3 fix).
func TestAccDbtCloudJobCIWithDeferringEnvironment(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// 3. CI job with deferring_environment_id (verify it's preserved after apply)
			{
				Config: testAccDbtCloudJobCIWithDeferringEnvironmentConfig(
					jobName, projectName, environmentName, "ci",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.ci_job", "deferring_environment_id"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "job_type", "ci"),
				),
			},
			// Apply again to verify deferring_environment_id is stable (not nulled out on re-apply)
			{
				Config: testAccDbtCloudJobCIWithDeferringEnvironmentConfig(
					jobName, projectName, environmentName, "ci",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.ci_job", "deferring_environment_id"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.ci_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
					"triggers.%",
					"triggers.custom_branch_only",
				},
			},
		},
	})
}

// TestAccDbtCloudJobMergeWithDeferringEnvironment tests that a Merge job correctly
// preserves deferring_environment_id after apply (Change 3 fix).
func TestAccDbtCloudJobMergeWithDeferringEnvironment(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// 4. Merge job with deferring_environment_id
			{
				Config: testAccDbtCloudJobCIWithDeferringEnvironmentConfig(
					jobName, projectName, environmentName, "merge",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.ci_job", "deferring_environment_id"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "job_type", "merge"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.ci_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
					"triggers.%",
					"triggers.custom_branch_only",
				},
			},
		},
	})
}

// TestAccDbtCloudJobCIWithCostOptimizationFeatures verifies that legacy SAO
// remains in Terraform state for a CI job even though it is intentionally
// suppressed at the API boundary to avoid a 405 response.
func TestAccDbtCloudJobCIWithCostOptimizationFeatures(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	jobNameUpdated := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Create: legacy SAO is retained in Terraform state without sending it to CI.
			{
				Config: testAccDbtCloudJobCIWithCostOptimizationFeaturesConfig(
					jobName, projectName, environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "job_type", "ci"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "cost_optimization_features.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_job.ci_job",
						"cost_optimization_features.*",
						"state_aware_orchestration",
					),
				),
			},
			// Rename only: legacy SAO must remain in Terraform state on an unrelated update.
			{
				Config: testAccDbtCloudJobCIWithCostOptimizationFeaturesConfig(
					jobNameUpdated, projectName, environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "name", jobNameUpdated),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "cost_optimization_features.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						"dbtcloud_job.ci_job",
						"cost_optimization_features.*",
						"state_aware_orchestration",
					),
				),
			},
		},
	})
}

func testAccDbtCloudJobCostOptimizationFeaturesConfig(
	jobName, projectName, environmentName, featuresVal string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name = "%s"
    dbt_version = "%s"
    type = "deployment"
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
  cost_optimization_features = %s
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, featuresVal)
}

func testAccDbtCloudJobCIWithDeferringEnvironmentConfig(
	jobName, projectName, environmentName, jobType string,
) string {
	githubWebhook := "false"
	gitProviderWebhook := "false"
	onMerge := "false"
	if jobType == "ci" {
		githubWebhook = "true"
		gitProviderWebhook = "true"
	} else if jobType == "merge" {
		onMerge = "true"
	}
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name = "%s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "ci_job" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment.environment_id
  deferring_environment_id = dbtcloud_environment.test_job_environment.environment_id
  execute_steps = [
    "dbt build -s state:modified+"
  ]
  triggers = {
    "github_webhook": %s,
    "git_provider_webhook": %s,
    "schedule": false,
    "on_merge": %s,
  }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, githubWebhook, gitProviderWebhook, onMerge)
}

func testAccDbtCloudJobCIWithCostOptimizationFeaturesConfig(
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
    type = "deployment"
}

resource "dbtcloud_job" "ci_job" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment.environment_id
  job_type = "ci"
  cost_optimization_features = ["state_aware_orchestration"]
  execute_steps = [
    "dbt build -s state:modified+"
  ]
  triggers = {
    "github_webhook": false,
    "git_provider_webhook": false,
    "schedule": false,
  }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName)
}

func testAccDbtCloudJobCIWithDbtStateConfig(
	jobName, projectName, environmentName, featuresVal string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name = "%s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "ci_job" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment.environment_id
  cost_optimization_features = %s
  execute_steps = [
    "dbt build -s state:modified+"
  ]
  triggers = {
    "github_webhook": true,
    "git_provider_webhook": false,
    "schedule": false,
  }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, featuresVal)
}

func testAccDbtCloudJobResourceIntervalCronConfig(
	jobName, projectName, environmentName string,
	interval int,
	days []int,
) string {
	daysStr := "["
	for i, day := range days {
		if i > 0 {
			daysStr += ","
		}
		daysStr += fmt.Sprintf("%d", day)
	}
	daysStr += "]"

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
    "schedule": true,
  }
  schedule_type = "interval_cron"
  schedule_interval = %d
  schedule_days = %s
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, interval, daysStr)
}

// TestAccDbtCloudJobResourceJobTypeInPlaceTransitions verifies that transitions
// between scheduled, other, and merge do NOT require resource replacement.
func TestAccDbtCloudJobResourceJobTypeInPlaceTransitions(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	var firstID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Create with job_type = "scheduled"
			{
				Config: testAccDbtCloudJobResourceJobTypeTransitionConfig(jobName, projectName, environmentName, "scheduled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "job_type", "scheduled"),
					// Capture the resource ID for later comparison
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["dbtcloud_job.test_job"]
						if !ok {
							return fmt.Errorf("resource dbtcloud_job.test_job not found")
						}
						firstID = rs.Primary.ID
						return nil
					},
				),
			},
			// Transition scheduled -> other: must be in-place (same ID)
			{
				Config: testAccDbtCloudJobResourceJobTypeTransitionConfig(jobName, projectName, environmentName, "other"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "job_type", "other"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["dbtcloud_job.test_job"]
						if !ok {
							return fmt.Errorf("resource dbtcloud_job.test_job not found")
						}
						if rs.Primary.ID != firstID {
							return fmt.Errorf("expected same resource ID after in-place job_type change (scheduled->other), got %s, want %s", rs.Primary.ID, firstID)
						}
						return nil
					},
				),
			},
			// Transition other -> merge: must be in-place (same ID)
			{
				Config: testAccDbtCloudJobResourceJobTypeTransitionConfig(jobName, projectName, environmentName, "merge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "job_type", "merge"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["dbtcloud_job.test_job"]
						if !ok {
							return fmt.Errorf("resource dbtcloud_job.test_job not found")
						}
						if rs.Primary.ID != firstID {
							return fmt.Errorf("expected same resource ID after in-place job_type change (other->merge), got %s, want %s", rs.Primary.ID, firstID)
						}
						return nil
					},
				),
			},
		},
	})
}

// TestAccDbtCloudJobResourceJobTypeCIRequiresReplacement verifies that a
// transition from ci to scheduled forces resource replacement (new ID).
func TestAccDbtCloudJobResourceJobTypeCIRequiresReplacement(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	var ciID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Create with job_type = "ci"
			{
				Config: testAccDbtCloudJobResourceJobTypeCIConfig(jobName, projectName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "job_type", "ci"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["dbtcloud_job.test_job"]
						if !ok {
							return fmt.Errorf("resource dbtcloud_job.test_job not found")
						}
						ciID = rs.Primary.ID
						return nil
					},
				),
			},
			// Transition ci -> scheduled: must create a new resource (different ID)
			{
				Config: testAccDbtCloudJobResourceJobTypeTransitionConfig(jobName, projectName, environmentName, "scheduled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "job_type", "scheduled"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["dbtcloud_job.test_job"]
						if !ok {
							return fmt.Errorf("resource dbtcloud_job.test_job not found")
						}
						if rs.Primary.ID == ciID {
							return fmt.Errorf("expected new resource ID after ci->scheduled job_type change (replacement required), but ID is unchanged: %s", ciID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccDbtCloudJobResourceJobTypeTransitionConfig(jobName, projectName, environmentName, jobType string) string {
	// The API infers job_type from triggers, so we must set matching triggers.
	scheduleTrigger := "false"
	onMergeTrigger := "false"
	switch jobType {
	case "scheduled":
		scheduleTrigger = "true"
	case "merge":
		onMergeTrigger = "true"
	}
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name       = "%s"
    dbt_version = "%s"
    type       = "deployment"
}

resource "dbtcloud_job" "test_job" {
    name           = "%s"
    project_id     = dbtcloud_project.test_job_project.id
    environment_id = dbtcloud_environment.test_job_environment.environment_id
    execute_steps  = ["dbt build"]
    triggers = {
        "github_webhook"      : false,
        "git_provider_webhook": false,
        "schedule"            : %s,
        "on_merge"            : %s,
    }
    job_type = "%s"
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, scheduleTrigger, onMergeTrigger, jobType)
}

func testAccDbtCloudJobResourceJobTypeCIConfig(jobName, projectName, environmentName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name       = "%s"
    dbt_version = "%s"
    type       = "deployment"
}

resource "dbtcloud_job" "test_job" {
    name                     = "%s"
    project_id               = dbtcloud_project.test_job_project.id
    environment_id           = dbtcloud_environment.test_job_environment.environment_id
    deferring_environment_id = dbtcloud_environment.test_job_environment.environment_id
    execute_steps            = ["dbt build"]
    triggers = {
        "github_webhook"      : false,
        "git_provider_webhook": false,
        "schedule"            : false,
    }
    job_type            = "ci"
    run_compare_changes = true
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName)
}
