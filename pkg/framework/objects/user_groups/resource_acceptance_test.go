package user_groups_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var userID = acctest_config.AcceptanceTestConfig.DbtCloudUserId
var groupIDs = fmt.Sprintf("[%s]", acctest_config.AcceptanceTestConfig.DbtCloudGroupIds)
var GroupName = acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

var addRoleTestStep = resource.TestStep{
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
}

var removeRoleTestStep = resource.TestStep{
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
		//we should only have 3 groups now that we check that there is no item at index 3 (starts at 0)
		// TODO: This test depends on the number of groups defined in the acceptance test config, which is not guaranteed.
		resource.TestCheckNoResourceAttr(
			"dbtcloud_user_groups.test_user_groups",
			"group_ids.3",
		),
	),
}

func TestAccDbtCloudUserGroupsResource(t *testing.T) {
	var importStateTestStep = resource.TestStep{
		ResourceName:            "dbtcloud_user_groups.test_user_groups",
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			addRoleTestStep,
			removeRoleTestStep,
			importStateTestStep,
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
