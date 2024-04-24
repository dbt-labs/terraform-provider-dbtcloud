package resources_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudEnvironmentVariableJobOverrideResource(t *testing.T) {

	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentVariableName := strings.ToUpper(
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
	)
	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentVariableJobOverrideValue := strings.ToUpper(
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
	)
	environmentVariableJobOverrideValueNew := strings.ToUpper(
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudEnvironmentVariableJobOverrideDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudEnvironmentVariableJobOverrideResourceBasicConfig(
					projectName,
					environmentName,
					environmentVariableName,
					jobName,
					environmentVariableJobOverrideValue,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentVariableJobOverrideExists(
						"dbtcloud_environment_variable_job_override.test_env_var_job_override",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						"name",
						fmt.Sprintf("DBT_%s", environmentVariableName),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable_job_override.test_env_var_job_override",
						"name",
						fmt.Sprintf("DBT_%s", environmentVariableName),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable_job_override.test_env_var_job_override",
						"raw_value",
						environmentVariableJobOverrideValue,
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_environment_variable_job_override.test_env_var_job_override",
						"environment_variable_job_override_id",
					),
				),
			},
			// RENAME
			// MODIFY
			{
				Config: testAccDbtCloudEnvironmentVariableJobOverrideResourceBasicConfig(
					projectName,
					environmentName,
					environmentVariableName,
					jobName,
					environmentVariableJobOverrideValueNew,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentVariableJobOverrideExists(
						"dbtcloud_environment_variable_job_override.test_env_var_job_override",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable_job_override.test_env_var_job_override",
						"name",
						fmt.Sprintf("DBT_%s", environmentVariableName),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable_job_override.test_env_var_job_override",
						"raw_value",
						environmentVariableJobOverrideValueNew,
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_environment_variable_job_override.test_env_var_job_override",
						"environment_variable_job_override_id",
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_environment_variable_job_override.test_env_var_job_override",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudEnvironmentVariableJobOverrideResourceBasicConfig(
	projectName, environmentName, environmentVariableName, jobName, environmentVariableJobOverrideValue string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_environment" "test_env" {
  name        = "%s"
  type = "deployment"
  dbt_version = "%s"
  project_id = dbtcloud_project.test_project.id
}

resource "dbtcloud_environment_variable" "test_env_var" {
  name        = "DBT_%s"
  project_id = dbtcloud_project.test_project.id
  environment_values = {
    "project": "Baa",
    "%s": "Moo"
  }
  depends_on = [
    dbtcloud_project.test_project,
    dbtcloud_environment.test_env
  ]
}

resource dbtcloud_job test_job_sched {
  environment_id = dbtcloud_environment.test_env.environment_id
  execute_steps = ["dbt test"]
  name = "%s"
  project_id = dbtcloud_project.test_project.id
  triggers = {
    "github_webhook" : false,
    "git_provider_webhook" : false,
    "schedule" : false
  }
  num_threads = 4
  schedule_days     = [0, 1, 2, 3, 4, 5, 6]
  schedule_type     = "days_of_week"
  schedule_interval = 6
}

resource dbtcloud_environment_variable_job_override test_env_var_job_override {
	job_definition_id = dbtcloud_job.test_job_sched.id
	project_id = dbtcloud_project.test_project.id
	name = dbtcloud_environment_variable.test_env_var.name
	raw_value = "%s"
}


`, projectName, environmentName, DBT_CLOUD_VERSION, environmentVariableName, environmentName, jobName, environmentVariableJobOverrideValue)
}

func testAccCheckDbtCloudEnvironmentVariableJobOverrideExists(
	resource string,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get Project ID")
		}
		jobID, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get Job ID")
		}
		envVarOverrideID, err := strconv.Atoi(
			strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[2],
		)
		if err != nil {
			return fmt.Errorf("Can't get the env var override ID")
		}

		_, err = apiClient.GetEnvironmentVariableJobOverride(
			projectId,
			jobID,
			envVarOverrideID,
		)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudEnvironmentVariableJobOverrideDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_environment_variable_job_override" {
			continue
		}

		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get Project ID")
		}
		jobID, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get Job ID")
		}
		envVarOverrideID, err := strconv.Atoi(
			strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[2],
		)
		if err != nil {
			return fmt.Errorf("Can't get the env var override ID")
		}

		_, err = apiClient.GetEnvironmentVariableJobOverride(
			projectId,
			jobID,
			envVarOverrideID,
		)
		if err == nil {
			return fmt.Errorf("Environment variable still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
