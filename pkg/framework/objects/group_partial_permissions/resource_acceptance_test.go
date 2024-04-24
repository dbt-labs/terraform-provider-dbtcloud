package group_partial_permissions_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudGroupPartialPermissionsResource(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	groupName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1. CREATE
			{
				Config: testAccDbtCloudGroupPartialPermissionsResourceCreate(
					projectName,
					groupName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"name",
						groupName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"sso_mapping_groups.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"group_permissions.#",
						"2",
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"group_permissions.*",
						map[string]string{
							"permission_set": "developer",
						},
					),
				),
			},
			// 2. ADD ANOTHER RESOURCE
			{
				Config: testAccDbtCloudGroupPartialPermissionsResourceAddResource(
					projectName,
					groupName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"name",
						groupName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"sso_mapping_groups.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"group_permissions.#",
						"2",
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"group_permissions.*",
						map[string]string{
							"permission_set": "developer",
						},
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"name",
						groupName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"sso_mapping_groups.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"group_permissions.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"group_permissions.0.permission_set",
						"admin",
					),
				),
			},
			// 3. MODIFYING EXISTING RESOURCE
			{
				Config: testAccDbtCloudGroupPartialPermissionsResourceModifyExisting(
					projectName,
					groupName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"name",
						groupName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"group_permissions.#",
						"1",
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"group_permissions.*",
						map[string]string{
							"permission_set": "developer",
						},
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"name",
						groupName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"sso_mapping_groups.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"group_permissions.#",
						"2",
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"group_permissions.*",
						map[string]string{
							"permission_set": "job_viewer",
						},
					),
				),
			},
			// 4. RENAME RESOURCE
			{
				Config: testAccDbtCloudGroupPartialPermissionsResourceModifyExisting(
					projectName,
					groupName+"2",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"name",
						groupName+"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"group_permissions.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission",
						"group_permissions.0.permission_set",
						"developer",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"name",
						groupName+"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"sso_mapping_groups.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"group_permissions.#",
						"2",
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"group_permissions.*",
						map[string]string{
							"permission_set": "admin",
						},
					),
				),
			},
			// 5. REMOVE ONE RESOURCE
			{
				Config: testAccDbtCloudGroupPartialPermissionsResourceRemoveResource(
					projectName,
					groupName+"2",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"name",
						groupName+"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"sso_mapping_groups.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"group_permissions.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group_partial_permissions.test_group_partial_permission2",
						"group_permissions.0.permission_set",
						"admin",
					),
				),
			},
		},
	})
}

func testAccDbtCloudGroupPartialPermissionsResourceCreate(projectName, groupName string) string {

	groupPartialPermissionConfig := fmt.Sprintf(`

resource "dbtcloud_project" "test_group_project" {
	name = "%s"
}

resource "dbtcloud_group_partial_permissions" "test_group_partial_permission" {
	name  				= "%s"
	sso_mapping_groups 	= ["group1", "group2"]
	group_permissions = [
		{
			permission_set 	= "developer"
			project_id    	= dbtcloud_project.test_group_project.id 
			all_projects  	= false
		},
		{
			permission_set 	= "analyst"
			project_id    	= dbtcloud_project.test_group_project.id 
			all_projects  	= false
		}
	]
}
`, projectName, groupName)
	return groupPartialPermissionConfig
}

func testAccDbtCloudGroupPartialPermissionsResourceAddResource(
	projectName, groupName string,
) string {

	groupPartialPermissionConfig := fmt.Sprintf(`

resource "dbtcloud_project" "test_group_project" {
	name = "%s"
}

resource "dbtcloud_group_partial_permissions" "test_group_partial_permission" {
	name  				= "%s"
	sso_mapping_groups 	= ["group1", "group2"]
	group_permissions = [
		{
			permission_set 	= "developer"
			project_id    	= dbtcloud_project.test_group_project.id 
			all_projects  	= false
		},
		{
			permission_set 	= "analyst"
			project_id    	= dbtcloud_project.test_group_project.id 
			all_projects  	= false
		}
	]
}

resource "dbtcloud_group_partial_permissions" "test_group_partial_permission2" {
	name  				= "%s"
	sso_mapping_groups 	= ["group1", "group2"]
	group_permissions = [
		{
			permission_set 	= "admin"
			project_id    	= dbtcloud_project.test_group_project.id 
			all_projects  	= false
		},
	]
	depends_on = [
		dbtcloud_group_partial_permissions.test_group_partial_permission
	  ]
}
`, projectName, groupName, groupName)
	return groupPartialPermissionConfig
}

func testAccDbtCloudGroupPartialPermissionsResourceModifyExisting(
	projectName, groupName string,
) string {

	groupPartialPermissionConfig := fmt.Sprintf(`

resource "dbtcloud_project" "test_group_project" {
	name = "%s"
}

resource "dbtcloud_group_partial_permissions" "test_group_partial_permission" {
	name  				= "%s"
	sso_mapping_groups 	= ["group1", "group2"]
	group_permissions = [
		{
			permission_set 	= "developer"
			project_id    	= dbtcloud_project.test_group_project.id 
			all_projects  	= false
		}
	]
}

resource "dbtcloud_group_partial_permissions" "test_group_partial_permission2" {
	name  				= "%s"
	sso_mapping_groups 	= ["group1", "group2"]
	group_permissions = [
		{
			permission_set 	= "admin"
			project_id    	= dbtcloud_project.test_group_project.id 
			all_projects  	= false
		},
		{
			permission_set 	= "job_viewer"
			all_projects  	= true
		},
	]
	depends_on = [
		dbtcloud_group_partial_permissions.test_group_partial_permission
	  ]
}
`, projectName, groupName, groupName)
	return groupPartialPermissionConfig
}

func testAccDbtCloudGroupPartialPermissionsResourceRemoveResource(
	projectName, groupName string,
) string {

	groupPartialPermissionConfig := fmt.Sprintf(`

resource "dbtcloud_project" "test_group_project" {
	name = "%s"
}

resource "dbtcloud_group_partial_permissions" "test_group_partial_permission2" {
	name  				= "%s"
	sso_mapping_groups 	= ["group1", "group2"]
	group_permissions = [
		{
			permission_set 	= "admin"
			project_id    	= dbtcloud_project.test_group_project.id 
			all_projects  	= false
		},
	]
}
`, projectName, groupName)
	return groupPartialPermissionConfig
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("DBT_CLOUD_ACCOUNT_ID"); v == "" {
		t.Fatal("DBT_CLOUD_ACCOUNT_ID must be set for acceptance tests")
	}
	if v := os.Getenv("DBT_CLOUD_TOKEN"); v == "" {
		t.Fatal("DBT_CLOUD_TOKEN must be set for acceptance tests")
	}
}
