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

// TestAccDbtCloudJobResourceScheduledToOther tests that a scheduled job
// can be changed to 'other' job_type in-place (allowed transition).
func TestAccDbtCloudJobResourceScheduledToOther(t *testing.T) {
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create scheduled job
			{
				Config: testAccDbtCloudJobResourceJobTypeConfig(
					jobName,
					projectName,
					environmentName,
					"scheduled",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "job_type", "scheduled"),
				),
			},
			// Step 2: Change to 'other' - should update in-place
			{
				Config: testAccDbtCloudJobResourceJobTypeConfig(
					jobName,
					projectName,
					environmentName,
					"other",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "job_type", "other"),
				),
			},
			// Step 3: Change back to 'scheduled' - should update in-place
			{
				Config: testAccDbtCloudJobResourceJobTypeConfig(
					jobName,
					projectName,
					environmentName,
					"scheduled",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "job_type", "scheduled"),
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

func testAccDbtCloudJobResourceJobTypeConfig(
	jobName, projectName, environmentName, jobType string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "project" {
    name = "%s"
}

resource "dbtcloud_environment" "deploy_env" {
    project_id = dbtcloud_project.project.id
    name = "%s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "test_job" {
    name = "%s"
    project_id = dbtcloud_project.project.id
    environment_id = dbtcloud_environment.deploy_env.environment_id
    execute_steps = ["dbt build"]
    job_type = "%s"
    
    triggers = {
        "github_webhook" : false
        "git_provider_webhook" : false
        "schedule" : true
        "on_merge" : false
    }
    
    schedule_type = "every_day"
    schedule_hours = [9]
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, jobType)
}

// TestAccDbtCloudJobResourceScheduledWithSAO tests a scheduled job with
// run_compare_changes (SAO) enabled alongside deferring_environment_id.
// This validates that SAO works correctly on scheduled jobs.
func TestAccDbtCloudJobResourceScheduledWithSAO(t *testing.T) {
	if acctest_config.IsDbtCloudPR() {
		// SAO (Advanced CI) requires account-level setting to be enabled
		t.Skip("Skipping: SAO requires Advanced CI to be enabled in account")
	}

	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceScheduledWithSAOConfig(
					jobName,
					projectName,
					environmentName,
					true, // SAO enabled
					"--select state:modified+",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.scheduled_sao_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.scheduled_sao_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.scheduled_sao_job", "run_compare_changes", "true"),
					resource.TestCheckResourceAttr("dbtcloud_job.scheduled_sao_job", "compare_changes_flags", "--select state:modified+"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.scheduled_sao_job", "deferring_environment_id"),
				),
			},
			// Step 2: Disable SAO
			{
				Config: testAccDbtCloudJobResourceScheduledWithSAOConfig(
					jobName,
					projectName,
					environmentName,
					false, // SAO disabled
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.scheduled_sao_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.scheduled_sao_job", "run_compare_changes", "false"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_job.scheduled_sao_job",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"validate_execute_steps",
				},
			},
		},
	})
}

func testAccDbtCloudJobResourceScheduledWithSAOConfig(
	jobName, projectName, environmentName string,
	saoEnabled bool,
	compareChangesFlags string,
) string {
	runCompareChanges := "false"
	compareChangesFlagsConfig := ""
	if saoEnabled {
		runCompareChanges = "true"
		if compareChangesFlags != "" {
			compareChangesFlagsConfig = fmt.Sprintf(`compare_changes_flags = "%s"`, compareChangesFlags)
		}
	}

	return fmt.Sprintf(`
resource "dbtcloud_project" "project" {
    name = "%s"
}

resource "dbtcloud_environment" "prod_env" {
    project_id = dbtcloud_project.project.id
    name = "PROD %s"
    dbt_version = "%s"
    type = "deployment"
    deployment_type = "production"
}

resource "dbtcloud_job" "scheduled_sao_job" {
    name = "%s"
    project_id = dbtcloud_project.project.id
    environment_id = dbtcloud_environment.prod_env.environment_id
    execute_steps = ["dbt build"]
    
    deferring_environment_id = dbtcloud_environment.prod_env.environment_id
    run_compare_changes = %s
    %s
    
    triggers = {
        "github_webhook" : false
        "git_provider_webhook" : false
        "schedule" : true
        "on_merge" : false
    }
    
    schedule_type = "every_day"
    schedule_hours = [9]
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, runCompareChanges, compareChangesFlagsConfig)
}

// TestAccDbtCloudJobResourceCompareChangesFlagsCustomSelector tests that
// compare_changes_flags works with a custom selector (not the default).
func TestAccDbtCloudJobResourceCompareChangesFlagsCustomSelector(t *testing.T) {
	if acctest_config.IsDbtCloudPR() {
		t.Skip("Skipping: run_compare_changes requires Advanced CI to be enabled in account")
	}

	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with custom selector
			{
				Config: testAccDbtCloudJobResourceCompareChangesFlagsConfig(
					jobName,
					projectName,
					environmentName,
					"--select state:modified+ --exclude tag:skip_ci",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "name", jobName),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "run_compare_changes", "true"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "compare_changes_flags", "--select state:modified+ --exclude tag:skip_ci"),
				),
			},
			// Step 2: Change selector
			{
				Config: testAccDbtCloudJobResourceCompareChangesFlagsConfig(
					jobName,
					projectName,
					environmentName,
					"--select state:new",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.ci_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.ci_job", "compare_changes_flags", "--select state:new"),
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

func testAccDbtCloudJobResourceCompareChangesFlagsConfig(
	jobName, projectName, environmentName, compareChangesFlags string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "project" {
    name = "%s"
}

resource "dbtcloud_environment" "prod_env" {
    project_id = dbtcloud_project.project.id
    name = "PROD %s"
    dbt_version = "%s"
    type = "deployment"
    deployment_type = "production"
}

resource "dbtcloud_environment" "ci_env" {
    project_id = dbtcloud_project.project.id
    name = "CI %s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "ci_job" {
    name = "%s"
    project_id = dbtcloud_project.project.id
    environment_id = dbtcloud_environment.ci_env.environment_id
    execute_steps = ["dbt build -s state:modified+ --fail-fast"]
    
    deferring_environment_id = dbtcloud_environment.prod_env.environment_id
    run_compare_changes = true
    compare_changes_flags = "%s"
    job_type = "ci"
    
    triggers = {
        "github_webhook" : true
        "git_provider_webhook" : true
        "schedule" : false
        "on_merge" : false
    }
}
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, environmentName, acctest_config.DBT_CLOUD_VERSION, jobName, compareChangesFlags)
}
