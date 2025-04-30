package partial_environment_variable_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudPartialEnvironmentVariableResource(t *testing.T) {
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	envVarName := "DBT_" + acctest.RandStringFromCharSet(10, acctest_config.CharSetAlphaUpper)
	environmentName1 := "development"
	environmentName2 := "production"
	envValue1 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	envValue2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	updatedEnvValue1 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentVariableDestroy,
		Steps: []resource.TestStep{
			// Initial creation with single environment value
			{
				Config: testAccDbtCloudEnvironmentVariableResourceBasicConfig(
					projectName,
					environmentName1,
					envVarName,
					envValue1,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentVariableExists(
						"dbtcloud_partial_environment_variable.test_env_var",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_partial_environment_variable.test_env_var",
						"name",
						envVarName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_partial_environment_variable.test_env_var",
						fmt.Sprintf("environment_values.%s", environmentName1),
						envValue1,
					),
				),
			},
			// Modify: update existing value and add new environment
			{
				Config: testAccDbtCloudEnvironmentVariableResourceMultipleConfig(
					projectName,
					environmentName1,
					environmentName2,
					envVarName,
					updatedEnvValue1,
					envValue2,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentVariableExists(
						"dbtcloud_partial_environment_variable.test_env_var",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_partial_environment_variable.test_env_var",
						"name",
						envVarName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_partial_environment_variable.test_env_var",
						fmt.Sprintf("environment_values.%s", environmentName1),
						updatedEnvValue1,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_partial_environment_variable.test_env_var",
						fmt.Sprintf("environment_values.%s", environmentName2),
						envValue2,
					),
				),
			},
			// Split into two partial resources
			{
				Config: testAccDbtCloudEnvironmentVariableSplitConfig(
					projectName,
					environmentName1,
					environmentName2,
					envVarName,
					updatedEnvValue1,
					envValue2,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentVariableExists(
						"dbtcloud_partial_environment_variable.test_env_var_1",
					),
					testAccCheckDbtCloudEnvironmentVariableExists(
						"dbtcloud_partial_environment_variable.test_env_var_2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_partial_environment_variable.test_env_var_1",
						"name",
						envVarName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_partial_environment_variable.test_env_var_1",
						fmt.Sprintf("environment_values.%s", environmentName1),
						updatedEnvValue1,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_partial_environment_variable.test_env_var_2",
						"name",
						envVarName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_partial_environment_variable.test_env_var_2",
						fmt.Sprintf("environment_values.%s", environmentName2),
						envValue2,
					),
				),
			},
		},
	})
}

func testAccDbtCloudEnvironmentVariableResourceBasicConfig(
	projectName string,
	environmentName string,
	envVarName string,
	envValue string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_environment" "test_env" {
  name        = "%s"
  type = "development"
  project_id = dbtcloud_project.test_project.id
}

resource "dbtcloud_partial_environment_variable" "test_env_var" {
  project_id = dbtcloud_project.test_project.id
  name       = "%s"
  environment_values = {
    (dbtcloud_environment.test_env.name) = "%s"
  }
}
`, projectName, environmentName, envVarName, envValue)
}

func testAccDbtCloudEnvironmentVariableResourceMultipleConfig(
	projectName string,
	environmentName1 string,
	environmentName2 string,
	envVarName string,
	envValue1 string,
	envValue2 string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_environment" "test_env" {
  name        = "%s"
  type = "development"
  project_id = dbtcloud_project.test_project.id
}

resource "dbtcloud_environment" "test_env_prod" {
  name        = "%s"
  type = "deployment"
  dbt_version = "latest"
  deployment_type = "production"
  project_id = dbtcloud_project.test_project.id
}

resource "dbtcloud_partial_environment_variable" "test_env_var" {
  project_id = dbtcloud_project.test_project.id
  name       = "%s"
  environment_values = {
    (dbtcloud_environment.test_env.name) = "%s"
    (dbtcloud_environment.test_env_prod.name) = "%s"
  }
}
`, projectName, environmentName1, environmentName2, envVarName, envValue1, envValue2)
}

func testAccDbtCloudEnvironmentVariableSplitConfig(
	projectName string,
	environmentName1 string,
	environmentName2 string,
	envVarName string,
	envValue1 string,
	envValue2 string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_environment" "test_env" {
  name        = "%s"
  type = "development"
  project_id = dbtcloud_project.test_project.id
}

resource "dbtcloud_environment" "test_env_prod" {
  name        = "%s"
  type = "deployment"
  dbt_version = "latest"
  deployment_type = "production"
  project_id = dbtcloud_project.test_project.id
}

resource "dbtcloud_partial_environment_variable" "test_env_var_1" {
  project_id = dbtcloud_project.test_project.id
  name       = "%s"
  environment_values = {
    (dbtcloud_environment.test_env.name) = "%s"
  }
}

resource "dbtcloud_partial_environment_variable" "test_env_var_2" {
  project_id = dbtcloud_project.test_project.id
  name       = "%s"
  environment_values = {
    (dbtcloud_environment.test_env_prod.name) = "%s"
  }
  depends_on = [dbtcloud_partial_environment_variable.test_env_var_1]
}
`, projectName, environmentName1, environmentName2, envVarName, envValue1, envVarName, envValue2)
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
		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}

		idParts := strings.Split(rs.Primary.ID, ":")
		if len(idParts) != 2 {
			return fmt.Errorf("Unexpected ID format: %s", rs.Primary.ID)
		}

		projectID, err := strconv.Atoi(idParts[0])
		if err != nil {
			return fmt.Errorf("Error converting project ID: %s", err)
		}

		name := idParts[1]
		_, err = apiClient.GetEnvironmentVariable(projectID, name)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudEnvironmentVariableDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_partial_environment_variable" {
			continue
		}

		idParts := strings.Split(rs.Primary.ID, ":")
		if len(idParts) != 2 {
			return fmt.Errorf("Unexpected ID format: %s", rs.Primary.ID)
		}

		projectID, err := strconv.Atoi(idParts[0])
		if err != nil {
			return fmt.Errorf("Error converting project ID: %s", err)
		}

		name := idParts[1]
		_, err = apiClient.GetEnvironmentVariable(projectID, name)
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

func testAccDbtCloudEnvironmentVariableImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return rs.Primary.ID, nil
	}
}
