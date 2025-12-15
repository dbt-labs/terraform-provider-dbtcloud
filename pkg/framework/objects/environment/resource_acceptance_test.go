package environment_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

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

	resource.ParallelTest(t, resource.TestCase{
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
  credential_id = dbtcloud_snowflake_credential.test_credential.credential_id
  deployment_type = "production"
}

resource "dbtcloud_snowflake_credential" "test_credential" {
  project_id  = dbtcloud_project.test_project.id
  auth_type   = "password"
  num_threads = 16
  schema      = "analytics"
  user        = "my_snowflake_user"
  password    = "my_snowflake_password"
}
  
`, projectName, environmentName, acctest_config.DBT_CLOUD_VERSION, customBranch, useCustomBranch)
}

// testing for the global connection use case where connection_id is added at the env level
func TestAccDbtCloudEnvironmentResourceConnection(t *testing.T) {
	dbtVersionLatest := "latest"
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
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

	resource.ParallelTest(t, resource.TestCase{
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
  credential_id = dbtcloud_snowflake_credential.test_credential.credential_id
  deployment_type = "production"
  connection_id = dbtcloud_global_connection.test2.id
  enable_model_query_history = true
}

resource "dbtcloud_snowflake_credential" "test_credential" {
  project_id  = dbtcloud_project.test_project.id
  auth_type   = "password"
  num_threads = 16
  schema      = "analytics"
  user        = "my_snowflake_user"
  password    = "my_snowflake_password"
}
  
`, projectName, environmentName, dbtVersion, customBranch, useCustomBranch)
}

// TestAccDbtCloudEnvironmentResourceVersionless tests the environment resource with dbt_version set to versionless
// This is a special case where if the dbt_version is set to `versionless`, the dbt Cloud API may return `latest`
func TestAccDbtCloudEnvironmentResourceVersionless(t *testing.T) {
	dbtVersionless := "versionless"
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
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

func TestAccDbtCloudEnvironmentResourceFusion(t *testing.T) {
	dbtVersionFusion := "latest"
	dbtMinorVersion := "1.5.0-latest"
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudEnvironmentResourceNoConnectionBasicConfig(
					projectName,
					environmentName,
					dbtMinorVersion,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"dbt_version",
						dbtMinorVersion,
					),
				),
			},
			{
				Config: testAccDbtCloudEnvironmentResourceNoConnectionBasicConfig(
					projectName,
					environmentName,
					dbtVersionFusion,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"dbt_version",
						dbtVersionFusion,
					),
				),
			},
		},
	})
}

// TestAccDbtCloudEnvironmentResourceCustomBranchValidation tests that the custom_branch
// and use_custom_branch fields are validated correctly. This addresses GitHub issue #574.
func TestAccDbtCloudEnvironmentResourceCustomBranchValidation(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test: custom_branch set but use_custom_branch is false - should error
			{
				Config: testAccDbtCloudEnvironmentResourceCustomBranchWithoutUseFlag(
					projectName,
					environmentName,
				),
				ExpectError: regexp.MustCompile("Inconsistent custom branch configuration"),
			},
			// Test: use_custom_branch is true but custom_branch is not set - should error
			{
				Config: testAccDbtCloudEnvironmentResourceUseCustomBranchWithoutBranch(
					projectName,
					environmentName,
				),
				ExpectError: regexp.MustCompile("Missing custom_branch"),
			},
		},
	})
}

func testAccDbtCloudEnvironmentResourceCustomBranchWithoutUseFlag(
	projectName, environmentName string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_environment" "test_env" {
  name              = "%s"
  type              = "development"
  project_id        = dbtcloud_project.test_project.id
  use_custom_branch = false
  custom_branch     = "my-custom-branch"
}
`, projectName, environmentName)
}

func testAccDbtCloudEnvironmentResourceUseCustomBranchWithoutBranch(
	projectName, environmentName string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_environment" "test_env" {
  name              = "%s"
  type              = "development"
  project_id        = dbtcloud_project.test_project.id
  use_custom_branch = true
}
`, projectName, environmentName)
}
