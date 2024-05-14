package resources_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudFabricConnectionResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	database := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	server := "example.com"
	port := 1337

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudFabricConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudFabricConnectionResourceBasicConfig(
					connectionName,
					projectName,
					database,
					server,
					port,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists(
						"dbtcloud_fabric_connection.test_connection",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_fabric_connection.test_connection",
						"name",
						connectionName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_fabric_connection.test_connection",
						"database",
						database,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_fabric_connection.test_connection",
						"server",
						server,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_fabric_connection.test_connection",
						"port",
						fmt.Sprintf("%d", port),
					),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudFabricConnectionResourceBasicConfig(
					connectionName2,
					projectName,
					database,
					server,
					port,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists(
						"dbtcloud_fabric_connection.test_connection",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_fabric_connection.test_connection",
						"name",
						connectionName2,
					),
				),
			},
			// MODIFY BY ADDING CONFIG
			{
				Config: testAccDbtCloudFabricConnectionResourceFullConfig(
					connectionName2,
					projectName,
					database,
					server,
					port,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists(
						"dbtcloud_fabric_connection.test_connection",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_fabric_connection.test_connection",
						"port",
						fmt.Sprintf("%d", port),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_fabric_connection.test_connection",
						"retries",
						"1234",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_fabric_connection.test_connection",
						"login_timeout",
						"2345",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_fabric_connection.test_connection",
						"query_timeout",
						"3456",
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_fabric_connection.test_connection",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudFabricConnectionResourceBasicConfig(
	connectionName, projectName, database, server string, port int,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_fabric_connection" "test_connection" {
	project_id = dbtcloud_project.test_project.id
	name = "%s"
	database = "%s"
	server = "%s"
	port = %d
}
`, projectName, connectionName, database, server, port)
}

func testAccDbtCloudFabricConnectionResourceFullConfig(
	connectionName, projectName, database, server string, port int,
) string {
	return fmt.Sprintf(`
	resource "dbtcloud_project" "test_project" {
		name        = "%s"
	  }
	  
	  resource "dbtcloud_fabric_connection" "test_connection" {
		  project_id = dbtcloud_project.test_project.id
		  name = "%s"
		  database = "%s"
		  server = "%s"
		  port = %d

		  retries = 1234
		  login_timeout = 2345
		  query_timeout = 3456
	  }
`, projectName, connectionName, database, server, port)
}

func testAccCheckDbtCloudFabricConnectionDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_fabric_connection" {
			continue
		}
		projectId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0]
		connectionId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1]

		_, err := apiClient.GetConnection(connectionId, projectId)
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
