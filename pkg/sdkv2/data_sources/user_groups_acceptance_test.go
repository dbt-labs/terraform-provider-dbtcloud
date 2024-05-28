package data_sources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDbtCloudUserGroupsDataSource(t *testing.T) {

	var userID int
	if isDbtCloudPR() {
		userID = 1
	} else if os.Getenv("CI") != "" {
		userID = 54461
	} else {
		userID = 32
	}

	config := userGroups(userID)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.dbtcloud_user_groups.test", "user_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_user_groups.test", "group_ids.0"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_user_groups.test", "group_ids.1"),
	)

	resource.ParallelTest(t, resource.TestCase{
		Providers: providers(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  check,
			},
		},
	})
}

func userGroups(userID int) string {
	return fmt.Sprintf(`
    data "dbtcloud_user_groups" "test" {
        user_id = %d
    }
    `, userID)
}
