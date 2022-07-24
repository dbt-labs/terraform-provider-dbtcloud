package resources_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudEnvironmentVariableResource(t *testing.T) {

	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentVariableName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudEnvironmentVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudEnvironmentVariableResourceBasicConfig(projectName, environmentName, environmentVariableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentVariableExists("dbt_cloud_environment_variable.test_env_var"),
					resource.TestCheckResourceAttr("dbt_cloud_environment_variable.test_env_var", "name", fmt.Sprintf("DBT_%s", environmentVariableName)),
				),
			},
			// RENAME
			// 			// MODIFY
			// 			{
			// 				Config: testAccDbtCloudEnvironmentResourceModifiedConfig(projectName, projectName2, environmentName2),
			// 				Check: resource.ComposeTestCheckFunc(
			// 					testAccCheckDbtCloudEnvironmentExists("dbt_cloud_environment.test_env"),
			// 					resource.TestCheckResourceAttr("dbt_cloud_environment.test_env", "name", environmentName2),
			// 					resource.TestCheckResourceAttr("dbt_cloud_environment.test_env", "dbt_version", "1.0.1"),
			// 				),
			// 			},
			// IMPORT
			{
				ResourceName:            "dbt_cloud_environment_variable.test_env_var",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudEnvironmentVariableResourceBasicConfig(projectName, environmentName, environmentVariableName string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "%s"
}

resource "dbt_cloud_environment" "test_env" {
  name        = "%s"
  type = "deployment"
  dbt_version = "1.0.0"
  project_id = dbt_cloud_project.test_project.id
}

resource "dbt_cloud_environment_variable" "test_env_var" {
  name        = "DBT_%s"
  project_id = dbt_cloud_project.test_project.id
  environment_values = {
    "%s": "Moo"
  }
}
`, projectName, environmentName, environmentVariableName, environmentName)
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
		if rs.Type != "dbt_cloud_environment_variable" {
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
		notFoundErr := "not found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
