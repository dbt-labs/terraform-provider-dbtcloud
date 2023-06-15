package resources_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudProjectConnectionResource(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudProjectConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudProjectConnectionResourceBasicConfig(projectName, connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProjectConnectionExists("dbt_cloud_project_connection.test_project_connection"),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbt_cloud_project_connection.test_project_connection",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// EMPTY
			{
				Config: testAccDbtCloudProjectConnectionResourceEmptyConfig(projectName, connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProjectConnectionEmpty("dbt_cloud_project.test_project"),
				),
			},
		},
	})
}

func testAccDbtCloudProjectConnectionResourceBasicConfig(projectName, connectionName string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "%s"
}

resource "dbt_cloud_connection" "test_connection" {
  name        = "%s"
  type = "snowflake"
  project_id = dbt_cloud_project.test_project.id
  account = "test"
  database = "db"
  warehouse = "wh"
  role = "user"
  allow_sso = false
  allow_keep_alive = false
}

resource "dbt_cloud_project_connection" "test_project_connection" {
  project_id = dbt_cloud_project.test_project.id
  connection_id = dbt_cloud_connection.test_connection.connection_id
}
`, projectName, connectionName)
}

func testAccDbtCloudProjectConnectionResourceEmptyConfig(projectName, connectionName string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "%s"
}

resource "dbt_cloud_connection" "test_connection" {
  name        = "%s"
  type = "snowflake"
  project_id = dbt_cloud_project.test_project.id
  account = "test"
  database = "db"
  warehouse = "wh"
  role = "user"
  allow_sso = false
  allow_keep_alive = false
}
`, projectName, connectionName)
}

func testAccCheckDbtCloudProjectConnectionExists(resource string) resource.TestCheckFunc {
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
		project, err := apiClient.GetProject(projectId)
		if err != nil {
			return fmt.Errorf("Can't get project")
		}
		if project.ConnectionID == nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudProjectConnectionEmpty(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		project, err := apiClient.GetProject(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get project")
		}
		if project.ConnectionID != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudProjectConnectionDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbt_cloud_project_connection" {
			continue
		}
		projectId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0]
		project, err := apiClient.GetProject(projectId)
		if project != nil {
			return fmt.Errorf("Project still exists")
		}
		notFoundErr := "not found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
