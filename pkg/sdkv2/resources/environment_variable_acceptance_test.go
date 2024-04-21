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

func TestAccDbtCloudEnvironmentVariableResource(t *testing.T) {

	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentVariableName := strings.ToUpper(
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudEnvironmentVariableDestroy,
		Steps: []resource.TestStep{
			// SECRET ENV VAR
			{
				Config: testAccDbtCloudEnvironmentVariableResourceBasicConfig(
					projectName,
					environmentName,
					fmt.Sprintf("ENV_SECRET_%s", environmentVariableName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentVariableExists(
						"dbtcloud_environment_variable.test_env_var",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						"name",
						fmt.Sprintf("DBT_ENV_SECRET_%s", environmentVariableName),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						"environment_values.%",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						"environment_values.project",
						"Baa",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						fmt.Sprintf("environment_values.%s", environmentName),
						"Moo",
					),
				),
			},
			// NON SECRET ENV VAR
			{
				Config: testAccDbtCloudEnvironmentVariableResourceBasicConfig(
					projectName,
					environmentName,
					environmentVariableName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentVariableExists(
						"dbtcloud_environment_variable.test_env_var",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						"name",
						fmt.Sprintf("DBT_%s", environmentVariableName),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						"environment_values.%",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						"environment_values.project",
						"Baa",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						fmt.Sprintf("environment_values.%s", environmentName),
						"Moo",
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudEnvironmentVariableResourceModifiedConfig(
					projectName,
					environmentName,
					environmentVariableName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentVariableExists(
						"dbtcloud_environment_variable.test_env_var",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						"name",
						fmt.Sprintf("DBT_%s", environmentVariableName),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						"environment_values.%",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						"environment_values.project",
						"Oink",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment_variable.test_env_var",
						fmt.Sprintf("environment_values.%s", environmentName),
						"Neigh",
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_environment_variable.test_env_var",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudEnvironmentVariableResourceBasicConfig(
	projectName, environmentName, environmentVariableName string,
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
`, projectName, environmentName, DBT_CLOUD_VERSION, environmentVariableName, environmentName)
}

func testAccDbtCloudEnvironmentVariableResourceModifiedConfig(
	projectName, environmentName, environmentVariableName string,
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
    "project": "Oink",
    "%s": "Neigh"
  }
  depends_on = [
    dbtcloud_project.test_project,
    dbtcloud_environment.test_env
  ]
}
`, projectName, environmentName, DBT_CLOUD_VERSION, environmentVariableName, environmentName)
}

func testAccCheckDbtCloudEnvironmentVariableExists(resource string) resource.TestCheckFunc {
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
			return fmt.Errorf("Can't get projectId")
		}

		environmentVariableName := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1]

		_, err = apiClient.GetEnvironmentVariable(projectId, environmentVariableName)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudEnvironmentVariableDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_environment_variable" {
			continue
		}
		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}

		environmentVariableName := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1]
		_, err = apiClient.GetEnvironmentVariable(projectId, environmentVariableName)
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
