package group_users_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudGroupUsersDataSource(t *testing.T) {

	groupName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := group_users(groupName)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet(
			"data.dbtcloud_group_users.test_group_users_read",
			"group_id",
		),
		// we check that there is no user in the group as we just created it
		resource.TestCheckResourceAttr(
			"data.dbtcloud_group_users.test_group_users_read",
			"users.#",
			"0",
		),
	)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  check,
			},
		},
	})
}

func group_users(groupName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_group" "test_group" {
    name = "%s"
}

data "dbtcloud_group_users" "test_group_users_read" {
    group_id = dbtcloud_group.test_group.id
}
`, groupName)
}
