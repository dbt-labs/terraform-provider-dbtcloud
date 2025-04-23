package resources_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudConnectionResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	oAuthClientID := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	oAuthClientSecret := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudConnectionResourceBasicConfig(
					connectionName,
					projectName,
					oAuthClientID,
					oAuthClientSecret,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists("dbtcloud_connection.test_connection"),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_connection",
						"name",
						connectionName,
					),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudConnectionResourceBasicConfig(
					connectionName2,
					projectName,
					oAuthClientID,
					oAuthClientSecret,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists("dbtcloud_connection.test_connection"),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_connection",
						"name",
						connectionName2,
					),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudConnectionResourceRedshiftConfig(
					connectionName,
					projectName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists(
						"dbtcloud_connection.test_redshift_connection",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_redshift_connection",
						"name",
						connectionName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_redshift_connection",
						"database",
						"db",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_redshift_connection",
						"port",
						"5432",
					),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudConnectionResourceRedshiftConfig(
					connectionName2,
					projectName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists(
						"dbtcloud_connection.test_redshift_connection",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_redshift_connection",
						"name",
						connectionName2,
					),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudConnectionResourcePostgresConfig(
					connectionName,
					projectName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists(
						"dbtcloud_connection.test_postgres_connection",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_postgres_connection",
						"name",
						connectionName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_postgres_connection",
						"database",
						"db",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_postgres_connection",
						"port",
						"5432",
					),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudConnectionResourcePostgresConfig(
					connectionName2,
					projectName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists(
						"dbtcloud_connection.test_postgres_connection",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_postgres_connection",
						"name",
						connectionName2,
					),
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

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	databricksHost := "databricks.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudConnectionResourceDatabricksConfig(
					connectionName,
					projectName,
					databricksHost,
					"/my/databricks",
					"moo",
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists(
						"dbtcloud_connection.test_databricks_connection",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"name",
						connectionName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"host_name",
						databricksHost,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"http_path",
						"/my/databricks",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"catalog",
						"moo",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_connection.test_databricks_connection",
						"adapter_id",
					),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudConnectionResourceDatabricksConfig(
					connectionName2,
					projectName,
					databricksHost,
					"/my/databricks",
					"moo",
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists(
						"dbtcloud_connection.test_databricks_connection",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"name",
						connectionName2,
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudConnectionResourceDatabricksConfig(
					connectionName2,
					projectName,
					databricksHost,
					"/my/databricks_new",
					"moo2",
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists(
						"dbtcloud_connection.test_databricks_connection",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"name",
						connectionName2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"http_path",
						"/my/databricks_new",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"catalog",
						"moo2",
					),
				),
			},
			// MODIFY TO AUTH
			{
				Config: testAccDbtCloudConnectionResourceDatabricksConfig(
					connectionName2,
					projectName,
					databricksHost,
					"/my/databricks_new",
					"moo2",
					true,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists(
						"dbtcloud_connection.test_databricks_connection",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"name",
						connectionName2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"http_path",
						"/my/databricks_new",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"catalog",
						"moo2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"oauth_client_id",
						"client",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_databricks_connection",
						"oauth_client_secret",
						"secret",
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_connection.test_databricks_connection",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth_client_id", "oauth_client_secret"},
			},
		},
	})
}

func TestAccDbtCloudConnectionPrivateLinkResource(t *testing.T) {

	// we only test this explicitly as we can't create a PL connection and need to read from existing ones
	if os.Getenv("DBT_ACCEPTANCE_TEST_PRIVATE_LINK") == "" {
		t.Skip("Skipping acceptance tests as DBT_ACCEPTANCE_TEST_PRIVATE_LINK is not set")
	}

	endpointName := os.Getenv("DBT_ACCEPTANCE_TEST_PRIVATE_LINK_NAME")
	endpointURL := os.Getenv("DBT_ACCEPTANCE_TEST_PRIVATE_LINK_URL")

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudConnectionResourcePrivateLinkConfig(
					connectionName,
					projectName,
					endpointName,
					endpointURL,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists("dbtcloud_connection.test_connection"),
					resource.TestCheckResourceAttr(
						"dbtcloud_connection.test_connection",
						"name",
						connectionName,
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_connection.test_connection",
						"private_link_endpoint_id",
					),
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

func testAccDbtCloudConnectionResourceBasicConfig(
	connectionName, projectName, oAuthClientID, oAuthClientSecret string,
) string {
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
  allow_sso = true
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

func testAccDbtCloudConnectionResourceDatabricksConfig(
	connectionName, projectName string,
	databricksHost string,
	httpPath string,
	catalog string,
	oAuth bool,
) string {

	oauthConfig := ""
	if oAuth {
		oauthConfig = `
		oauth_client_id = "client"
		oauth_client_secret = "secret"	
		`
	}

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
  http_path  = "%s"
  catalog    = "%s"
  %s
}
`, projectName, connectionName, databricksHost, httpPath, catalog, oauthConfig)
}

func testAccDbtCloudConnectionResourcePrivateLinkConfig(
	connectionName, projectName, endpointName, endpointURL string,
) string {
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
		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		projectId, connectionId, err := helper.SplitIDToStrings(
			rs.Primary.ID,
			"dbtcloud_connection",
		)
		if err != nil {
			return err
		}

		_, err = apiClient.GetConnection(connectionId, projectId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudConnectionDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_connection" {
			continue
		}
		projectId, connectionId, err := helper.SplitIDToStrings(
			rs.Primary.ID,
			"dbtcloud_connection",
		)
		if err != nil {
			return err
		}

		_, err = apiClient.GetConnection(connectionId, projectId)
		if err == nil {
			return fmt.Errorf("Connection still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
