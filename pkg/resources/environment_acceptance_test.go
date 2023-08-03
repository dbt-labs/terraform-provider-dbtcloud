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

func TestAccDbtCloudEnvironmentResource(t *testing.T) {

	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudEnvironmentResourceBasicConfig(projectName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "name", environmentName),
					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "deployment_type", "production"),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudEnvironmentResourceBasicConfig(projectName, environmentName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "name", environmentName2),
				),
			},
			// MODIFY ADDING CRED
			{
				Config: testAccDbtCloudEnvironmentResourceModifiedConfig(projectName, environmentName2, "", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "name", environmentName2),
					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "dbt_version", "1.0.1"),
					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "custom_branch", ""),
					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "use_custom_branch", "false"),
					resource.TestCheckResourceAttrSet("dbtcloud_environment.test_env", "credential_id"),
					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "deployment_type", ""),
				),
			},
			// MODIFY CUSTOM BRANCH
			{
				Config: testAccDbtCloudEnvironmentResourceModifiedConfig(projectName, environmentName2, "main", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "name", environmentName2),
					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "dbt_version", "1.0.1"),
					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "custom_branch", "main"),
					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "use_custom_branch", "true"),
					resource.TestCheckResourceAttrSet("dbtcloud_environment.test_env", "credential_id"),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_environment.test_env",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudEnvironmentResourceBasicConfig(projectName, environmentName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_environment" "test_env" {
  name        = "%s"
  type = "deployment"
  dbt_version = "1.0.1"
  project_id = dbtcloud_project.test_project.id
  deployment_type = "production"
}
`, projectName, environmentName)
}

func testAccDbtCloudEnvironmentResourceModifiedConfig(projectName, environmentName, customBranch, useCustomBranch string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_environment" "test_env" {
  name        = "%s"
  type = "deployment"
  dbt_version = "1.0.1"
  custom_branch = "%s"
  use_custom_branch = %s
  project_id = dbtcloud_project.test_project.id
  credential_id = dbtcloud_bigquery_credential.test_credential.credential_id
}

resource "dbtcloud_bigquery_credential" "test_credential" {
	project_id  = dbtcloud_project.test_project.id
	dataset     = "my_bq_dataset"
	num_threads = 16
  }
  
`, projectName, environmentName, customBranch, useCustomBranch)
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
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}

		environmentId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get environmentId")
		}

		_, err = apiClient.GetEnvironment(projectId, environmentId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudEnvironmentDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_environment" {
			continue
		}
		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectId")
		}

		environmentId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1])
		if err != nil {
			return fmt.Errorf("Can't get environmentId")
		}
		_, err = apiClient.GetEnvironment(projectId, environmentId)
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
