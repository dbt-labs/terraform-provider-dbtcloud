package group_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudGroupDataSource(t *testing.T) {

	groupName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := group(groupName)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbtcloud_group.test_group_read", "name", groupName),
		resource.TestCheckResourceAttrSet("data.dbtcloud_group.test_group_read", "is_active"),
		resource.TestCheckResourceAttrSet(
			"data.dbtcloud_group.test_group_read",
			"assign_by_default",
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

func group(groupName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_group" "test_group" {
    name = "%s"
}

data "dbtcloud_group" "test_group_read" {
    group_id = dbtcloud_group.test_group.id
}
`, groupName)
}
