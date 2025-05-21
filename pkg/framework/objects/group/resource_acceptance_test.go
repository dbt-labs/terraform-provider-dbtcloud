package group_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudGroupResource(t *testing.T) {

	groupName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	groupName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudGroupResourceBasicConfig(groupName, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudGroupExists("dbtcloud_group.test_group"),
					resource.TestCheckResourceAttr("dbtcloud_group.test_group", "name", groupName),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"group_permissions.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"group_permissions.1.permission_set",
						"member",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"group_permissions.1.all_projects",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"group_permissions.0.permission_set",
						"developer",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"group_permissions.0.all_projects",
						"false",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_group.test_group",
						"group_permissions.0.project_id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"sso_mapping_groups.0",
						"group1",
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudGroupResourceFullConfig(groupName2, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudGroupExists("dbtcloud_group.test_group"),
					resource.TestCheckResourceAttr("dbtcloud_group.test_group", "name", groupName2),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"assign_by_default",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"group_permissions.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"group_permissions.0.permission_set",
						"member",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"group_permissions.0.all_projects",
						"false",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_group.test_group",
						"group_permissions.0.project_id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"group_permissions.1.all_projects",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"group_permissions.1.permission_set",
						"developer",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_group.test_group",
						"sso_mapping_groups.#",
						"2",
					),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_group.test_group",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDbtCloudGroupResourceBasicConfig(groupName, projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_group" "test_group" {
    name = "%s"
    group_permissions {
        permission_set = "member"
        all_projects = true
    }
    group_permissions {
        permission_set = "developer"
        all_projects = false
        project_id = dbtcloud_project.test_project.id
		writable_environment_categories = ["production", "other"]
    }
	sso_mapping_groups = ["group1"]
}
`, projectName, groupName)
}

func testAccDbtCloudGroupResourceFullConfig(groupName, projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_group" "test_group" {
    name = "%s"
    assign_by_default = true
    group_permissions {
        permission_set = "member"
        all_projects = false
        project_id = dbtcloud_project.test_project.id
    }
    group_permissions {
        permission_set = "developer"
        all_projects = true
		writable_environment_categories = ["development", "staging"]
    }
	sso_mapping_groups = ["group1", "group2"]
}
`, projectName, groupName)
}

func testAccCheckDbtCloudGroupExists(resource string) resource.TestCheckFunc {
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
		groupID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get groupID")
		}
		_, err = apiClient.GetGroup(groupID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudGroupDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_group" {
			continue
		}
		groupID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get groupID")
		}
		_, err = apiClient.GetGroup(groupID)
		if err == nil {
			return fmt.Errorf("Group still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
