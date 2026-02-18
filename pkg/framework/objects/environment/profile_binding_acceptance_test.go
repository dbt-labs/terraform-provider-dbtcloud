package environment_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudEnvironmentResourcePrimaryProfile(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	profileKey := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			// CREATE with primary_profile_id
			{
				Config: testAccDbtCloudEnvironmentWithProfileConfig(
					projectName,
					environmentName,
					profileKey,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttr(
						"dbtcloud_environment.test_env",
						"name",
						environmentName,
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_environment.test_env",
						"primary_profile_id",
					),
				),
			},
			// UPDATE - remove primary_profile_id
			{
				Config: testAccDbtCloudEnvironmentWithoutProfileConfig(
					projectName,
					environmentName,
					profileKey,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckNoResourceAttr(
						"dbtcloud_environment.test_env",
						"primary_profile_id",
					),
				),
			},
		},
	})
}

func testAccDbtCloudEnvironmentWithProfileConfig(
	projectName, environmentName, profileKey string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_global_connection" "test_connection" {
  name = "env_profile_test_connection_%s"

  snowflake = {
    account   = "test-account"
    warehouse = "test-warehouse"
    database  = "test-database"
  }
}

resource "dbtcloud_snowflake_credential" "test_credential" {
  is_active   = true
  project_id  = dbtcloud_project.test_project.id
  auth_type   = "password"
  database    = "test-database"
  role        = "test-role"
  warehouse   = "test-warehouse"
  schema      = "test_schema"
  user        = "test-user"
  password    = "test-password"
  num_threads = 3
}

resource "dbtcloud_profile" "test_profile" {
  project_id     = dbtcloud_project.test_project.id
  key            = "%s"
  connection_id  = dbtcloud_global_connection.test_connection.id
  credentials_id = dbtcloud_snowflake_credential.test_credential.credential_id
}

resource "dbtcloud_environment" "test_env" {
  name               = "%s"
  type               = "deployment"
  dbt_version        = "latest"
  project_id         = dbtcloud_project.test_project.id
  deployment_type    = "production"
  primary_profile_id = dbtcloud_profile.test_profile.profile_id
}
`, projectName, projectName, profileKey, environmentName)
}

// TestAccDbtCloudEnvironmentResourceDirectFKsToProfile tests switching an
// environment from direct FK configuration (connection_id + credential_id)
// to profile-based configuration (primary_profile_id).
func TestAccDbtCloudEnvironmentResourceDirectFKsToProfile(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	profileKey := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			// CREATE with direct FKs
			{
				Config: testAccDbtCloudEnvironmentWithoutProfileConfig(
					projectName,
					environmentName,
					profileKey,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_environment.test_env",
						"connection_id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_environment.test_env",
						"credential_id",
					),
					resource.TestCheckNoResourceAttr(
						"dbtcloud_environment.test_env",
						"primary_profile_id",
					),
				),
			},
			// UPDATE - switch to profile
			{
				Config: testAccDbtCloudEnvironmentWithProfileConfig(
					projectName,
					environmentName,
					profileKey,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_environment.test_env",
						"primary_profile_id",
					),
				),
			},
		},
	})
}

// TestAccDbtCloudEnvironmentResourceProfileChange tests changing an
// environment from one profile to a different profile.
func TestAccDbtCloudEnvironmentResourceProfileChange(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	profileKeyA := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	profileKeyB := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			// CREATE with profile A (backed by conn_a)
			{
				Config: testAccDbtCloudEnvironmentWithTwoProfiles(
					projectName,
					environmentName,
					profileKeyA,
					profileKeyB,
					"a",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttrPair(
						"dbtcloud_environment.test_env", "primary_profile_id",
						"dbtcloud_profile.profile_a", "profile_id",
					),
					resource.TestCheckResourceAttrPair(
						"dbtcloud_environment.test_env", "connection_id",
						"dbtcloud_global_connection.conn_a", "id",
					),
				),
			},
			// UPDATE - switch to profile B (backed by conn_b)
			{
				Config: testAccDbtCloudEnvironmentWithTwoProfiles(
					projectName,
					environmentName,
					profileKeyA,
					profileKeyB,
					"b",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttrPair(
						"dbtcloud_environment.test_env", "primary_profile_id",
						"dbtcloud_profile.profile_b", "profile_id",
					),
					resource.TestCheckResourceAttrPair(
						"dbtcloud_environment.test_env", "connection_id",
						"dbtcloud_global_connection.conn_b", "id",
					),
				),
			},
		},
	})
}

// TestAccDbtCloudEnvironmentResourceProfileImport tests that importing an
// environment that uses a profile binding works correctly.
func TestAccDbtCloudEnvironmentResourceProfileImport(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	profileKey := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudEnvironmentDestroy,
		Steps: []resource.TestStep{
			// CREATE with primary_profile_id
			{
				Config: testAccDbtCloudEnvironmentWithProfileConfig(
					projectName,
					environmentName,
					profileKey,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudEnvironmentExists("dbtcloud_environment.test_env"),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_environment.test_env",
						"primary_profile_id",
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_environment.test_env",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"primary_profile_id"},
			},
		},
	})
}

// testAccDbtCloudEnvironmentWithTwoProfiles creates infrastructure with two
// profiles and assigns one of them (selected by activeProfile: "a" or "b")
// to the environment.
func testAccDbtCloudEnvironmentWithTwoProfiles(
	projectName, environmentName, profileKeyA, profileKeyB, activeProfile string,
) string {
	profileRef := "dbtcloud_profile.profile_a.profile_id"
	if activeProfile == "b" {
		profileRef = "dbtcloud_profile.profile_b.profile_id"
	}

	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_global_connection" "conn_a" {
  name = "env_profile_change_conn_a_%s"

  snowflake = {
    account   = "test-account"
    warehouse = "test-warehouse"
    database  = "test-database"
  }
}

resource "dbtcloud_global_connection" "conn_b" {
  name = "env_profile_change_conn_b_%s"

  snowflake = {
    account   = "test-account"
    warehouse = "test-warehouse"
    database  = "test-database"
  }
}

resource "dbtcloud_snowflake_credential" "test_credential" {
  is_active   = true
  project_id  = dbtcloud_project.test_project.id
  auth_type   = "password"
  database    = "test-database"
  role        = "test-role"
  warehouse   = "test-warehouse"
  schema      = "test_schema"
  user        = "test-user"
  password    = "test-password"
  num_threads = 3
}

resource "dbtcloud_profile" "profile_a" {
  project_id     = dbtcloud_project.test_project.id
  key            = "%s"
  connection_id  = dbtcloud_global_connection.conn_a.id
  credentials_id = dbtcloud_snowflake_credential.test_credential.credential_id
}

resource "dbtcloud_profile" "profile_b" {
  project_id     = dbtcloud_project.test_project.id
  key            = "%s"
  connection_id  = dbtcloud_global_connection.conn_b.id
  credentials_id = dbtcloud_snowflake_credential.test_credential.credential_id
}

resource "dbtcloud_environment" "test_env" {
  name               = "%s"
  type               = "deployment"
  dbt_version        = "latest"
  project_id         = dbtcloud_project.test_project.id
  deployment_type    = "production"
  primary_profile_id = %s
}
`, projectName, projectName, projectName, profileKeyA, profileKeyB, environmentName, profileRef)
}

func testAccDbtCloudEnvironmentWithoutProfileConfig(
	projectName, environmentName, profileKey string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_global_connection" "test_connection" {
  name = "env_profile_test_connection_%s"

  snowflake = {
    account   = "test-account"
    warehouse = "test-warehouse"
    database  = "test-database"
  }
}

resource "dbtcloud_snowflake_credential" "test_credential" {
  is_active   = true
  project_id  = dbtcloud_project.test_project.id
  auth_type   = "password"
  database    = "test-database"
  role        = "test-role"
  warehouse   = "test-warehouse"
  schema      = "test_schema"
  user        = "test-user"
  password    = "test-password"
  num_threads = 3
}

resource "dbtcloud_profile" "test_profile" {
  project_id     = dbtcloud_project.test_project.id
  key            = "%s"
  connection_id  = dbtcloud_global_connection.test_connection.id
  credentials_id = dbtcloud_snowflake_credential.test_credential.credential_id
}

resource "dbtcloud_environment" "test_env" {
  name            = "%s"
  type            = "deployment"
  dbt_version     = "latest"
  project_id      = dbtcloud_project.test_project.id
  deployment_type = "production"
  connection_id   = dbtcloud_global_connection.test_connection.id
  credential_id   = dbtcloud_snowflake_credential.test_credential.credential_id
}
`, projectName, projectName, profileKey, environmentName)
}
