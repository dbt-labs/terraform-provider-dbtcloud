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

func TestConformanceBasicConfig(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest_helper.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			acctest_helper.MakeExternalProviderTestStep(getBasicConfigTestStep(projectName, environmentName), acctest_config.LAST_VERSION_BEFORE_FRAMEWORK_MIGRATION),
			acctest_helper.MakeCurrentProviderNoOpTestStep(getBasicConfigTestStep(projectName, environmentName)),
		},
	})
}

func TestConformanceModifyConfig(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest_helper.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			acctest_helper.MakeExternalProviderTestStep(getBasicConfigWithModifiedConfigTestStep(projectName, environmentName, "", "false"), acctest_config.LAST_VERSION_BEFORE_FRAMEWORK_MIGRATION),
			acctest_helper.MakeCurrentProviderNoOpTestStep(getBasicConfigWithModifiedConfigTestStep(projectName, environmentName, "", "false")),
		},
	})
}

func TestConformanceModifyConfigCustomBranch(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest_helper.TestAccPreCheck(t) },
		CheckDestroy: testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			acctest_helper.MakeExternalProviderTestStep(getBasicConfigWithModifiedConfigTestStep(projectName, environmentName, "main", "true"), acctest_config.LAST_VERSION_BEFORE_FRAMEWORK_MIGRATION),
			acctest_helper.MakeCurrentProviderNoOpTestStep(getBasicConfigWithModifiedConfigTestStep(projectName, environmentName, "main", "true")),
		},
	})
}

func getBasicConfigTestStep(projectName, envName string) resource.TestStep {
	return resource.TestStep{
		Config: testAccDbtCloudEnvironmentResourceNoConnectionBasicConfig(
			projectName,
			envName,
			acctest_config.DBT_CLOUD_VERSION,
		),
		Check: resource.ComposeTestCheckFunc(
			testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
			resource.TestCheckResourceAttr(
				"dbtcloud_environment.test_env",
				"name",
				envName,
			),
		),
	}
}

func getBasicConfigWithModifiedConfigTestStep(projectName, environmentName, custom_branch, use_custom_branch string) resource.TestStep {
	return resource.TestStep{
		Config: testAccDbtCloudEnvironmentResourceNoConnectionModifiedConfig(
			projectName,
			environmentName,
			custom_branch,
			use_custom_branch,
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
				"dbt_version",
				acctest_config.DBT_CLOUD_VERSION,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_environment.test_env",
				"custom_branch",
				custom_branch,
			),
			resource.TestCheckResourceAttr(
				"dbtcloud_environment.test_env",
				"use_custom_branch",
				use_custom_branch,
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
	}
}

func getImportConfigTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:            "dbtcloud_environment.test_env",
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{},
	}
}

// testing for the historical use case where connection_id is not configured at the env level
func TestAccDbtCloudEnvironmentResourceNoConnection(t *testing.T) {
	initialEnvName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	newEnvName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			getBasicConfigTestStep(projectName, initialEnvName),
			getBasicConfigTestStep(projectName, newEnvName),
			getBasicConfigWithModifiedConfigTestStep(projectName, newEnvName, "", "false"),
			getBasicConfigWithModifiedConfigTestStep(projectName, newEnvName, "main", "true"),
			getImportConfigTestStep(),
		},
	})
}

func testAccDbtCloudEnvironmentResourceNoConnectionBasicConfig(
	projectName, environmentName, dbtVersion string,
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
`, projectName, environmentName, dbtVersion)
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
  
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, customBranch, useCustomBranch)
}

// testing for the global connection use case where connection_id is added at the env level
func TestAccDbtCloudEnvironmentResourceConnection(t *testing.T) {
	dbtVersionLatest := "latest"
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudEnvironmentResourceConnectionBasicConfig(
					projectName,
					environmentName,
					dbtVersionLatest,
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
					dbtVersionLatest,
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
			// MODIFY CUSTOM BRANCH
			{
				Config: testAccDbtCloudEnvironmentResourceConnectionModifiedConfig(
					projectName,
					environmentName2,
					"main",
					"true",
					dbtVersionLatest,
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
						dbtVersionLatest,
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
				ImportStateId:     fmt.Sprintf("%s:%s", "dbtcloud_project.test_project.id", "dbtcloud_environment.test_env.id"),
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					projectID := s.RootModule().Resources["dbtcloud_project.test_project"].Primary.ID
					environmentID := s.RootModule().Resources["dbtcloud_environment.test_env"].Primary.Attributes["environment_id"]
					return fmt.Sprintf("%s:%s", projectID, environmentID), nil
				},
				// TODO: Once the connection_id is mandatory, we can remove this exception and the custom logic for reading connection_id in the resource
				ImportStateVerifyIgnore: []string{"connection_id"},
			},
		},
	})
}

// TestAccDbtCloudEnvironmentResourceProjectUpdate tests that environments are not incorrectly updated when
// there is an update made to a project. Specifically tests the issue linked below where a project update
// would cascade a connection_id to all environments in the project.
// https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/334
func TestAccDbtCloudEnvironmentResourceProjectUpdate(t *testing.T) {
	dbtVersionLatest := "latest"
	environmentNameDev := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentNameProd := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectDescription := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectDescription2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudEnvironmentResourceDualConnectionConfig(
					projectName,
					projectDescription,
					environmentNameDev,
					environmentNameProd,
					dbtVersionLatest,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.dev"),
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.prod"),
				),
			},
			// MODIFY PROJECT
			{
				Config: testAccDbtCloudEnvironmentResourceDualConnectionConfig(
					projectName,
					projectDescription2,
					environmentNameDev,
					environmentNameProd,
					dbtVersionLatest,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.dev"),
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.prod"),
				),
			},
		},
	})
}

func testAccDbtCloudEnvironmentResourceDualConnectionConfig(
	projectName, projectDescription, environmentNameDev, environmentNameProd, dbtVersion string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
  description = "%s"
}

resource dbtcloud_global_connection dev {
  name = "test connection dev"

  snowflake = {
    account = "test"
    role = "role"
    warehouse = "warehouse"
    database = "database"
    allow_sso = false
  }
}

resource dbtcloud_global_connection prod {
  name = "test connection prod"

  snowflake = {
    account = "test"
    role = "role"
    warehouse = "warehouse"
    database = "database"
    allow_sso = false
  }
}

resource "dbtcloud_environment" "dev" {
  name        = "%s dev"
  type = "development"
  dbt_version = "%s"
  project_id = dbtcloud_project.test_project.id
  connection_id = dbtcloud_global_connection.dev.id
}

resource "dbtcloud_environment" "prod" {
  name        = "%s prod"
  type = "deployment"
  dbt_version = "%s"
  project_id = dbtcloud_project.test_project.id
  deployment_type = "production"
  connection_id = dbtcloud_global_connection.prod.id
}
  
  `, projectName, projectDescription, environmentNameDev, dbtVersion, environmentNameProd, dbtVersion)
}

func testAccDbtCloudEnvironmentResourceConnectionBasicConfig(
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
  deployment_type = "production"
  connection_id = dbtcloud_global_connection.test.id
  }
  
  `, projectName, environmentName, dbtVersion)
}

func testAccDbtCloudEnvironmentResourceConnectionModifiedConfig(
	projectName, environmentName, customBranch, useCustomBranch, dbtVersion string,
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
  enable_model_query_history = true
}

resource "dbtcloud_bigquery_credential" "test_credential" {
	project_id  = dbtcloud_project.test_project.id
	dataset     = "my_bq_dataset"
	num_threads = 16
  }
  
`, projectName, environmentName, dbtVersion, customBranch, useCustomBranch)
}

// TestAccDbtCloudEnvironmentResourceVersionless tests the environment resource with dbt_version set to versionless
// This is a special case where if the dbt_version is set to `versionless`, the dbt Cloud API may return `latest`
func TestAccDbtCloudEnvironmentResourceVersionless(t *testing.T) {
	dbtVersionless := "versionless"
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudEnvironmentResourceNoConnectionBasicConfig(
					projectName,
					environmentName,
					dbtVersionless,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestMatchResourceAttr(
						"dbtcloud_environment.test_env",
						"dbt_version",
						regexp.MustCompile("^versionless|latest$"),
					),
				),
			},
		},
	})
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
