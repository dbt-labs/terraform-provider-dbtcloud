package environment_test

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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// These tests were added to validate SIGN-136
func TestAccDbtCloudEnvironmentResourceNoDeploymentType(t *testing.T) {
	initialEnvName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			getNoDeploymentTestStep(projectName, initialEnvName),
		},
	})
}

func getNoDeploymentTestStep(projectName, envName string) resource.TestStep {
	return resource.TestStep{
		Config: testAccDbtCloudEnvironmentResourceNoDeploymentType(
			projectName,
			envName,
			acctest_config.DBT_CLOUD_VERSION,
		),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
			resource.TestCheckNoResourceAttr(
				"dbtcloud_environment.test_env",
				"deployment_type",
			),
		),
	}
}

func TestAccDbtCloudEnvironmentResourceAllOptionalParams(t *testing.T) {
	initialEnvName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			getAllOptionalTestStep(projectName, initialEnvName),
		},
	})
}

func getAllOptionalTestStep(projectName, envName string) resource.TestStep {

	customBranch := "test_branch"
	deploymentType := "production"

	return resource.TestStep{
		Config: testAccDbtCloudEnvironmentResourceAllOptional(
			projectName,
			envName,
			acctest_config.DBT_CLOUD_VERSION,
			customBranch,
			deploymentType,
		),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
			resource.TestCheckResourceAttr(
				"dbtcloud_environment.test_env",
				"deployment_type",
				deploymentType,
			),

			resource.ComposeTestCheckFunc(
				testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
				resource.TestCheckResourceAttr(
					"dbtcloud_environment.test_env",
					"custom_branch",
					customBranch,
				),
			),

			resource.ComposeTestCheckFunc(
				testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
				resource.TestCheckResourceAttrSet(
					"dbtcloud_environment.test_env",
					"extended_attributes_id",
				),
			),
		),
	}
}

func testAccDbtCloudEnvironmentResourceNoDeploymentType(
	projectName, environmentName, dbtVersion string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource dbtcloud_global_connection test {
  name = "test connection"

  snowflake = {
    account = "test"
    role = "role"
    warehouse = "warehouse"
    database = "database"
    allow_sso = false
  }
}

resource "dbtcloud_environment" "test_env" {
  name        = "%s"
  type = "deployment"
  dbt_version = "%s"
  project_id = dbtcloud_project.test_project.id
  connection_id = dbtcloud_global_connection.test.id
  }
  
  `, projectName, environmentName, dbtVersion)
}

func testAccDbtCloudEnvironmentResourceAllOptional(
	projectName, environmentName, dbtVersion, customBranch, deploymentType string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource dbtcloud_global_connection test {
  name = "test connection"

  snowflake = {
    account = "test"
    role = "role"
    warehouse = "warehouse"
    database = "database"
    allow_sso = false,
  }
}

resource "dbtcloud_extended_attributes" "test" {
  project_id = dbtcloud_project.test_project.id
  extended_attributes = jsonencode({
    "key1": "value1",
    "key2": "value2",
  })
}

resource "dbtcloud_environment" "test_env" {
  name        = "%s"
  type = "deployment"
  dbt_version = "%s"
  project_id = dbtcloud_project.test_project.id
  connection_id = dbtcloud_global_connection.test.id

  custom_branch = "%s"
  deployment_type = "%s"

  extended_attributes_id = dbtcloud_extended_attributes.test.extended_attributes_id
}
  
  `, projectName, environmentName, dbtVersion, customBranch, deploymentType)
}

func testAccCheckDbtCloudEnvironmentExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}
		environmentId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get environmentId")
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}

		_, err = apiClient.GetEnvironment(projectId, environmentId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudEnvironmentDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_environment" {
			continue
		}

		// Get the project ID from the state
		projectID := rs.Primary.Attributes["project_id"]
		if projectID == "" {
			return fmt.Errorf("No project_id found in state")
		}

		// Get the environment ID from the state
		environmentID := rs.Primary.Attributes["environment_id"]
		if environmentID == "" {
			return fmt.Errorf("No environment_id found in state")
		}

		// Convert IDs to integers
		projectIDInt, err := strconv.Atoi(projectID)
		if err != nil {
			return fmt.Errorf("Error converting project_id to integer: %s", err)
		}

		environmentIDInt, err := strconv.Atoi(environmentID)
		if err != nil {
			return fmt.Errorf("Error converting environment_id to integer: %s", err)
		}

		_, err = apiClient.GetEnvironment(projectIDInt, environmentIDInt)
		if err == nil {
			return fmt.Errorf("Environment still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
