package resources_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testing for the historical use case where connection_id is not configured at the env level
func TestAccDbtCloudEnvironmentResourceNoConnection(t *testing.T) {

	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudEnvironmentResourceNoConnectionBasicConfig(
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"name",
						environmentName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"deployment_type",
						"production",
					),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudEnvironmentResourceNoConnectionBasicConfig(
					projectName,
					environmentName2,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"name",
						environmentName2,
					),
				),
			},
			// MODIFY ADDING CRED
			{
				Config: testAccDbtCloudEnvironmentResourceNoConnectionModifiedConfig(
					projectName,
					environmentName2,
					"",
					"false",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"name",
						environmentName2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"dbt_version",
						DBT_CLOUD_VERSION,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"custom_branch",
						"",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"use_custom_branch",
						"false",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_environment.test_env",
						"credential_id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"deployment_type",
						"production",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"connection_id",
						"0",
					),
				),
			},
			// MODIFY CUSTOM BRANCH
			{
				Config: testAccDbtCloudEnvironmentResourceNoConnectionModifiedConfig(
					projectName,
					environmentName2,
					"main",
					"true",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"name",
						environmentName2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"dbt_version",
						DBT_CLOUD_VERSION,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"custom_branch",
						"main",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"use_custom_branch",
						"true",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_environment.test_env",
						"credential_id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"connection_id",
						"0",
					),
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

func testAccDbtCloudEnvironmentResourceNoConnectionBasicConfig(
	projectName, environmentName string,
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
  deployment_type = "production"
}
`, projectName, environmentName, DBT_CLOUD_VERSION)
}

func testAccDbtCloudEnvironmentResourceNoConnectionModifiedConfig(
	projectName, environmentName, customBranch, useCustomBranch string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_environment" "test_env" {
  name        = "%s"
  type = "deployment"
  dbt_version = "%s"
  custom_branch = "%s"
  use_custom_branch = %s
  project_id = dbtcloud_project.test_project.id
  credential_id = dbtcloud_bigquery_credential.test_credential.credential_id
  deployment_type = "production"
}

resource "dbtcloud_bigquery_credential" "test_credential" {
	project_id  = dbtcloud_project.test_project.id
	dataset     = "my_bq_dataset"
	num_threads = 16
  }
  
`, projectName, environmentName, DBT_CLOUD_VERSION, customBranch, useCustomBranch)
}

// testing for the global connection use case where connection_id is added at the env level
func TestAccDbtCloudEnvironmentResourceConnection(t *testing.T) {
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudEnvironmentResourceConnectionBasicConfig(
					projectName,
					environmentName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"name",
						environmentName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"deployment_type",
						"production",
					),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudEnvironmentResourceConnectionBasicConfig(
					projectName,
					environmentName2,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"name",
						environmentName2,
					),
				),
			},
			// MODIFY ADDING CRED
			{
				Config: testAccDbtCloudEnvironmentResourceConnectionModifiedConfig(
					projectName,
					environmentName2,
					"",
					"false",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"name",
						environmentName2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"dbt_version",
						DBT_CLOUD_VERSION,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"custom_branch",
						"",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"use_custom_branch",
						"false",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_environment.test_env",
						"credential_id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"deployment_type",
						"production",
					),
				),
			},
			// MODIFY CUSTOM BRANCH
			{
				Config: testAccDbtCloudEnvironmentResourceConnectionModifiedConfig(
					projectName,
					environmentName2,
					"main",
					"true",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"name",
						environmentName2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"dbt_version",
						DBT_CLOUD_VERSION,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"custom_branch",
						"main",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"use_custom_branch",
						"true",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_environment.test_env",
						"credential_id",
					),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_environment.test_env",
				ImportState:       true,
				ImportStateVerify: true,
				// TODO: Once the connection_id is mandatory, we can remove this exception and the custom logic for reading connection_id in the resource
				ImportStateVerifyIgnore: []string{"connection_id"},
			},
		},
	})
}

func testAccDbtCloudEnvironmentResourceConnectionBasicConfig(
	projectName, environmentName string,
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
  deployment_type = "production"
  connection_id = dbtcloud_global_connection.test.id
  }
  
  `, projectName, environmentName, DBT_CLOUD_VERSION)
}

func testAccDbtCloudEnvironmentResourceConnectionModifiedConfig(
	projectName, environmentName, customBranch, useCustomBranch string,
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

resource dbtcloud_global_connection test2 {
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
  custom_branch = "%s"
  use_custom_branch = %s
  project_id = dbtcloud_project.test_project.id
  credential_id = dbtcloud_bigquery_credential.test_credential.credential_id
  deployment_type = "production"
  connection_id = dbtcloud_global_connection.test2.id
}

resource "dbtcloud_bigquery_credential" "test_credential" {
	project_id  = dbtcloud_project.test_project.id
	dataset     = "my_bq_dataset"
	num_threads = 16
  }
  
`, projectName, environmentName, DBT_CLOUD_VERSION, customBranch, useCustomBranch)
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
		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
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
