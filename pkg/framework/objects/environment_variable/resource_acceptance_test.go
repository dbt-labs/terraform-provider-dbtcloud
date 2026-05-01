package environment_variable_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func getTestInputData() (string, string, string) {
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentVariableName := strings.ToUpper(
		acctest.RandStringFromCharSet(10, acctest.CharSetAlpha),
	)
	return projectName, environmentName, environmentVariableName
}

func TestAccDbtCloudEnvironmentVariableResource(t *testing.T) {

	projectName, environmentName, environmentVariableName := getTestInputData()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentVariableDestroy,
		Steps: []resource.TestStep{
			getNonSecretEnvTestStep(projectName, environmentName, environmentVariableName),
			getImportTestStep(),
		},
	})
}

func TestAccDbtCloudEnvironmentVariableResourceSecret(t *testing.T) {

	projectName, environmentName, environmentVariableName := getTestInputData()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentVariableDestroy,
		Steps: []resource.TestStep{
			getSecretEnvTestStep(projectName, environmentName, environmentVariableName),
			getImportTestStepSecret(),
		},
	})
}

func TestAccDbtCloudEnvironmentVariableResourceModify(t *testing.T) {

	projectName, environmentName, environmentVariableName := getTestInputData()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentVariableDestroy,
		Steps: []resource.TestStep{
			getModifyConfigTestStep(projectName, environmentName, environmentVariableName),
			getImportTestStep(),
		},
	})
}

func getImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      "dbtcloud_environment_variable.test_env_var",
		ImportState:       true,
		ImportStateVerify: true,
		// environment_values is now readable from the API for non-secret vars.
	}
}

// getImportTestStepSecret ignores environment_values on import because the API
// masks secret values and cannot round-trip them.
func getImportTestStepSecret() resource.TestStep {
	return resource.TestStep{
		ResourceName:            "dbtcloud_environment_variable.test_env_var",
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"environment_values"},
	}
}

func getModifyConfigTestStep(projectName, environmentName, environmentVariableName string) resource.TestStep {
	return resource.TestStep{
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
	}
}

func getSecretEnvTestStep(projectName, environmentName, environmentVariableName string) resource.TestStep {
	return resource.TestStep{
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
	}
}

func getNonSecretEnvTestStep(projectName, environmentName, environmentVariableName string) resource.TestStep {
	return resource.TestStep{
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
	}
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
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, environmentVariableName, environmentName)
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
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, environmentVariableName, environmentName)
}

// TestAccDbtCloudEnvironmentVariableResourceDrift verifies that changes made to a
// non-secret env var outside of Terraform (e.g. via the dbt Cloud UI) are detected
// as drift on the next plan.
func TestAccDbtCloudEnvironmentVariableResourceDrift(t *testing.T) {
	projectName, environmentName, environmentVariableName := getTestInputData()

	var capturedProjectID int
	var capturedEnvVarName string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentVariableDestroy,
		Steps: []resource.TestStep{
			// Step 1: create the resource and capture its project ID and name for use in PreConfig.
			{
				Config: testAccDbtCloudEnvironmentVariableResourceBasicConfig(
					projectName,
					environmentName,
					environmentVariableName,
				),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["dbtcloud_environment_variable.test_env_var"]
						parts := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)
						var err error
						capturedProjectID, err = strconv.Atoi(parts[0])
						if err != nil {
							return fmt.Errorf("could not parse project ID from %s", rs.Primary.ID)
						}
						capturedEnvVarName = parts[1]
						return nil
					},
				),
			},
			// Step 2: change values out-of-band via the API, then assert the plan is non-empty
			// (drift detected). Applying the step also restores the original values.
			{
				PreConfig: func() {
					client, err := acctest_helper.SharedClient()
					if err != nil {
						panic(fmt.Sprintf("could not get shared client: %s", err))
					}
					envVar, err := client.GetEnvironmentVariable(capturedProjectID, capturedEnvVarName)
					if err != nil {
						panic(fmt.Sprintf("could not read env var for out-of-band change: %s", err))
					}
					envValuesMap := make(map[string]string)
					for _, v := range envVar.EnvironmentNameValues {
						envValuesMap[strconv.Itoa(v.ID)] = "OutOfBandValue"
					}
					if _, err := client.UpdateEnvironmentVariable(
						capturedProjectID,
						dbt_cloud.AbstractedEnvironmentVariable{
							Name:              capturedEnvVarName,
							ProjectID:         capturedProjectID,
							EnvironmentValues: envValuesMap,
						},
					); err != nil {
						panic(fmt.Sprintf("out-of-band env var update failed: %s", err))
					}
				},
				Config: testAccDbtCloudEnvironmentVariableResourceBasicConfig(
					projectName,
					environmentName,
					environmentVariableName,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
		},
	})
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
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

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
