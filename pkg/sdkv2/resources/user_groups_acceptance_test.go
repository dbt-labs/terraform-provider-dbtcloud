package resources_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudUserGroupsResource(t *testing.T) {

	var userID int
	var groupIDs string
	if value := os.Getenv("CI"); value != "" {
		userID = 54461
		groupIDs = "[531585, 531584, 531583]"
	} else {
		userID = 4
		groupIDs = "[104, 105, 106]"
	}

	GroupName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudUserGroupsResourceAddRole(userID, GroupName, groupIDs),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_user_groups.test_user_groups",
						"user_id",
						strconv.Itoa(userID),
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_user_groups.test_user_groups",
						"group_ids.0",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_user_groups.test_user_groups",
						"group_ids.3",
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudUserGroupsResourceRemoveRole(userID, GroupName, groupIDs),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_user_groups.test_user_groups",
						"user_id",
						strconv.Itoa(userID),
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_user_groups.test_user_groups",
						"group_ids.0",
					),
					// we should only have 3 groups now that we check that there is no item at index 3 (starts at 0)
					resource.TestCheckNoResourceAttr(
						"dbtcloud_user_groups.test_user_groups",
						"group_ids.3",
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_user_groups.test_user_groups",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudUserGroupsResourceAddRole(
	userID int,
	groupName string,
	groupIDs string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_group" "test_group" {
	name = "%s"
	group_permissions {
		permission_set = "member"
		all_projects   = true
	}
}

locals {
	original_groups = %s
	new_groups = concat(local.original_groups, [dbtcloud_group.test_group.id])
}

resource "dbtcloud_user_groups" "test_user_groups" {
	user_id = %d
	group_ids = local.new_groups
  }
`, groupName, groupIDs, userID)
}

func testAccDbtCloudUserGroupsResourceRemoveRole(
	userID int,
	groupName string,
	groupIDs string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_group" "test_group" {
	name = "%s"
	group_permissions {
		permission_set = "member"
		all_projects   = true
	}
}

resource "dbtcloud_user_groups" "test_user_groups" {
	user_id = %d
	group_ids = %s
  }
`, groupName, userID, groupIDs)
}
