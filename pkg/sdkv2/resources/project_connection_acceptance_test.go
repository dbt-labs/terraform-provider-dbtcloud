package resources_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudProjectConnectionResource(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudProjectConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudProjectConnectionResourceBasicConfig(
					projectName,
					connectionName,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudProjectConnectionExists(
						"dbtcloud_project_connection.test_project_connection",
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_project_connection.test_project_connection",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudProjectConnectionResourceBasicConfig(
	projectName, connectionName string,
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
  allow_sso = false
  allow_keep_alive = false
}

resource "dbtcloud_project_connection" "test_project_connection" {
  project_id = dbtcloud_project.test_project.id
  connection_id = dbtcloud_connection.test_connection.connection_id
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
		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
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

func testAccCheckDbtCloudProjectConnectionDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_project_connection" {
			continue
		}
		projectId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0]
		project, err := apiClient.GetProject(projectId)
		if project != nil {
			return fmt.Errorf("Project still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
