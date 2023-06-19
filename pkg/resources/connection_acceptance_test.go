package resources_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudConnectionResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	oAuthClientID := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	oAuthClientSecret := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudConnectionResourceBasicConfig(connectionName, projectName, oAuthClientID, oAuthClientSecret),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists("dbtcloud_connection.test_connection"),
					resource.TestCheckResourceAttr("dbtcloud_connection.test_connection", "name", connectionName),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudConnectionResourceBasicConfig(connectionName2, projectName, oAuthClientID, oAuthClientSecret),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists("dbtcloud_connection.test_connection"),
					resource.TestCheckResourceAttr("dbtcloud_connection.test_connection", "name", connectionName2),
				),
			},
			// 			// MODIFY
			// 			{
			// 				Config: testAccDbtCloudEnvironmentResourceModifiedConfig(projectName, projectName2, environmentName2),
			// 				Check: resource.ComposeTestCheckFunc(
			// 					testAccCheckDbtCloudProjectExists("dbtcloud_environment.test_env"),
			// 					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "name", environmentName2),
			// 					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "dbt_version", "1.0.1"),
			// 				),
			// 			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_connection.test_connection",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth_client_id", "oauth_client_secret"},
			},
		},
	})
}

func TestAccDbtCloudRedshiftConnectionResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudConnectionResourceRedshiftConfig(connectionName, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists("dbtcloud_connection.test_redshift_connection"),
					resource.TestCheckResourceAttr("dbtcloud_connection.test_redshift_connection", "name", connectionName),
					resource.TestCheckResourceAttr("dbtcloud_connection.test_redshift_connection", "database", "db"),
					resource.TestCheckResourceAttr("dbtcloud_connection.test_redshift_connection", "port", "5432"),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudConnectionResourceRedshiftConfig(connectionName2, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists("dbtcloud_connection.test_redshift_connection"),
					resource.TestCheckResourceAttr("dbtcloud_connection.test_redshift_connection", "name", connectionName2),
				),
			},
			// 			// MODIFY
			// 			{
			// 				Config: testAccDbtCloudEnvironmentResourceModifiedConfig(projectName, projectName2, environmentName2),
			// 				Check: resource.ComposeTestCheckFunc(
			// 					testAccCheckDbtCloudProjectExists("dbtcloud_environment.test_env"),
			// 					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "name", environmentName2),
			// 					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "dbt_version", "1.0.1"),
			// 				),
			// 			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_connection.test_redshift_connection",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth_client_id", "oauth_client_secret"},
			},
		},
	})
}

func TestAccDbtCloudPostgresConnectionResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudConnectionResourcePostgresConfig(connectionName, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists("dbtcloud_connection.test_postgres_connection"),
					resource.TestCheckResourceAttr("dbtcloud_connection.test_postgres_connection", "name", connectionName),
					resource.TestCheckResourceAttr("dbtcloud_connection.test_postgres_connection", "database", "db"),
					resource.TestCheckResourceAttr("dbtcloud_connection.test_postgres_connection", "port", "5432"),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudConnectionResourcePostgresConfig(connectionName2, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists("dbtcloud_connection.test_postgres_connection"),
					resource.TestCheckResourceAttr("dbtcloud_connection.test_postgres_connection", "name", connectionName2),
				),
			},
			// 			// MODIFY
			// 			{
			// 				Config: testAccDbtCloudEnvironmentResourceModifiedConfig(projectName, projectName2, environmentName2),
			// 				Check: resource.ComposeTestCheckFunc(
			// 					testAccCheckDbtCloudProjectExists("dbtcloud_environment.test_env"),
			// 					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "name", environmentName2),
			// 					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "dbt_version", "1.0.1"),
			// 				),
			// 			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_connection.test_postgres_connection",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth_client_id", "oauth_client_secret"},
			},
		},
	})
}

func TestAccDbtCloudDatabricksConnectionResource(t *testing.T) {

	testDatabricks := os.Getenv("TEST_DATABRICKS")
	if testDatabricks == "true" {
		connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
		connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
		projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
		databricksHost := "test.cloud.databricks.com"

		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckDbtCloudConnectionDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccDbtCloudConnectionResourceDatabricksConfig(connectionName, projectName, databricksHost),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckDbtCloudConnectionExists("dbtcloud_connection.test_databricks_connection"),
						resource.TestCheckResourceAttr("dbtcloud_connection.test_databricks_connection", "name", connectionName),
						resource.TestCheckResourceAttr("dbtcloud_connection.test_databricks_connection", "host_name", databricksHost),
						resource.TestCheckResourceAttr("dbtcloud_connection.test_databricks_connection", "http_path", "/my/databricks"),
						resource.TestCheckResourceAttr("dbtcloud_connection.test_databricks_connection", "catalog", "moo"),
						resource.TestCheckResourceAttrSet("dbtcloud_connection.test_databricks_connection", "adapter_id"),
					),
				},
				// RENAME
				{
					Config: testAccDbtCloudConnectionResourceDatabricksConfig(connectionName2, projectName, databricksHost),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckDbtCloudConnectionExists("dbtcloud_connection.test_databricks_connection"),
						resource.TestCheckResourceAttr("dbtcloud_connection.test_databricks_connection", "name", connectionName2),
					),
				},
				// 			// MODIFY
				// 			{
				// 				Config: testAccDbtCloudEnvironmentResourceModifiedConfig(projectName, projectName2, environmentName2),
				// 				Check: resource.ComposeTestCheckFunc(
				// 					testAccCheckDbtCloudProjectExists("dbtcloud_environment.test_env"),
				// 					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "name", environmentName2),
				// 					resource.TestCheckResourceAttr("dbtcloud_environment.test_env", "dbt_version", "1.0.1"),
				// 				),
				// 			},
				// IMPORT
				{
					ResourceName:            "dbtcloud_connection.test_databricks_connection",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{},
				},
			},
		})
	}
}

func TestAccDbtCloudConnectionPrivateLinkResource(t *testing.T) {

	// we only test this explicitly as we can't create a PL connection and need to read from existing ones
	if os.Getenv("DBT_ACCEPTANCE_TEST_PRIVATE_LINK") != "" {

		endpointName := os.Getenv("DBT_ACCEPTANCE_TEST_PRIVATE_LINK_NAME")
		endpointURL := os.Getenv("DBT_ACCEPTANCE_TEST_PRIVATE_LINK_URL")

		connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
		projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckDbtCloudConnectionDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccDbtCloudConnectionResourcePrivateLinkConfig(connectionName, projectName, endpointName, endpointURL),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckDbtCloudConnectionExists("dbtcloud_connection.test_connection"),
						resource.TestCheckResourceAttr("dbtcloud_connection.test_connection", "name", connectionName),
						resource.TestCheckResourceAttrSet("dbtcloud_connection.test_connection", "private_link_endpoint_id"),
					),
				},
				// IMPORT
				{
					ResourceName:            "dbtcloud_connection.test_connection",
					ImportState:             true,
					ImportStateVerifyIgnore: []string{},
				},
			},
		})
	}
}

func testAccDbtCloudConnectionResourceBasicConfig(connectionName, projectName, oAuthClientID, oAuthClientSecret string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_connection" "test_connection" {
  name        = "%s"
  type = "snowflake"
  project_id = dbtcloud_project.test_project.id
  account = "test"
  database = "db"
  warehouse = "wh"
  role = "user"
  allow_sso = false
  allow_keep_alive = false
  oauth_client_id = "%s"
  oauth_client_secret = "%s"
}
`, projectName, connectionName, oAuthClientID, oAuthClientSecret)
}

func testAccDbtCloudConnectionResourceRedshiftConfig(connectionName, projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_connection" "test_redshift_connection" {
  name        = "%s"
  type = "redshift"
  project_id = dbtcloud_project.test_project.id
  host_name = "test_host_name"
  database = "db"
  port = 5432
  tunnel_enabled = true
}
`, projectName, connectionName)
}

func testAccDbtCloudConnectionResourcePostgresConfig(connectionName, projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_connection" "test_postgres_connection" {
  name        = "%s"
  type = "postgres"
  project_id = dbtcloud_project.test_project.id
  host_name = "test_postgres"
  database = "db"
  port = 5432
  tunnel_enabled = true
}
`, projectName, connectionName)
}

func testAccDbtCloudConnectionResourceDatabricksConfig(connectionName, projectName string, databricksHost string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_connection" "test_databricks_connection" {
  name       = "%s"
  type       = "adapter"
  database   = ""
  project_id = dbtcloud_project.test_project.id
  host_name  = "%s"
  http_path  = "/my/databricks"
  catalog    = "moo"
}
`, projectName, connectionName, databricksHost)
}

func testAccDbtCloudConnectionResourcePrivateLinkConfig(connectionName, projectName, endpointName, endpointURL string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

data "dbtcloud_privatelink_endpoint" "test" {
  name = "%s"
  private_link_endpoint_url = "%s"
}

resource "dbtcloud_connection" "test_connection" {
  name        = "%s"
  type = "snowflake"
  project_id = dbtcloud_project.test_project.id
  account = "test"
  database = "db"
  warehouse = "wh"
  role = "user"
  allow_sso = false
  allow_keep_alive = false
  private_link_endpoint_id = data.dbtcloud_privatelink_endpoint.test.id
}
`, projectName, endpointName, endpointURL, connectionName)
}

//
// func testAccDbtCloudEnvironmentResourceModifiedConfig(projectName, projectName2, environmentName string) string {
// 	return fmt.Sprintf(`
// resource "dbtcloud_project" "test_project" {
//   name        = "%s"
// }
//
// resource "dbtcloud_project" "test_project_2" {
//   name        = "%s"
// }
//
// resource "dbtcloud_environment" "test_env" {
//   name        = "%s"
//   type = "deployment"
//   dbt_version = "1.0.1"
//   project_id = dbtcloud_project.test_project_2.id
// }
// `, projectName, projectName2, environmentName)
// }

func testAccCheckDbtCloudConnectionExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		projectId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0]
		connectionId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1]

		_, err := apiClient.GetConnection(connectionId, projectId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudConnectionDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_connection" {
			continue
		}
		projectId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0]
		connectionId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1]

		_, err := apiClient.GetConnection(connectionId, projectId)
		if err == nil {
			return fmt.Errorf("Connection still exists")
		}
		notFoundErr := "not found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
