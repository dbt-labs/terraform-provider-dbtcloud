package scim_group_permissions_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDbtCloudScimGroupPermissionsResource(t *testing.T) {
	groupName := "Test SCIM Group"
	projectName := "Test Project for SCIM Group"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudScimGroupPermissionsResourceBasicConfig(groupName, projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("dbtcloud_scim_group_permissions.test", "id"),
					resource.TestCheckResourceAttrSet("dbtcloud_scim_group_permissions.test", "group_id"),
					resource.TestCheckResourceAttr("dbtcloud_scim_group_permissions.test", "permissions.#", "1"),
				),
			},
			// Test updating permissions
			{
				Config: testAccDbtCloudScimGroupPermissionsResourceUpdatedConfig(groupName, projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("dbtcloud_scim_group_permissions.test", "id"),
					resource.TestCheckResourceAttrSet("dbtcloud_scim_group_permissions.test", "group_id"),
					resource.TestCheckResourceAttr("dbtcloud_scim_group_permissions.test", "permissions.#", "2"),
				),
			},
			// Test import
			{
				ResourceName:      "dbtcloud_scim_group_permissions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDbtCloudScimGroupPermissionsResourceBasicConfig(groupName, projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_group" "test_group" {
  name = "%s"
  assign_by_default = false
}

resource "dbtcloud_scim_group_permissions" "test" {
  group_id = dbtcloud_group.test_group.id
  
  permissions = [
    {
      permission_set = "developer"
      project_id     = dbtcloud_project.test_project.id
      all_projects   = false
      writable_environment_categories = ["development"]
    }
  ]
}
`, projectName, groupName)
}

func testAccDbtCloudScimGroupPermissionsResourceUpdatedConfig(groupName, projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_group" "test_group" {
  name = "%s"
  assign_by_default = false
}

resource "dbtcloud_scim_group_permissions" "test" {
  group_id = dbtcloud_group.test_group.id
  
  permissions = [
    {
      permission_set = "developer"
      project_id     = dbtcloud_project.test_project.id
      all_projects   = false
      writable_environment_categories = ["development"]
    },
    {
      permission_set = "analyst"
      project_id     = dbtcloud_project.test_project.id
      all_projects   = false
      writable_environment_categories = ["development", "staging"]
    }
  ]
}
`, projectName, groupName)
}
