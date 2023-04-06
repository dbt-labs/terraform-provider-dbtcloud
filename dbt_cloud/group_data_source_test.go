package dbt_cloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudGroupDataSource(t *testing.T) {

	groupName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := group(groupName)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbt_cloud_group.test_group_read", "name", groupName),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_group.test_group_read", "is_active"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_group.test_group_read", "assign_by_default"),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
resource "dbt_cloud_group" "test_group" {
    name = "%s"
}

data "dbt_cloud_group" "test_group_read" {
    group_id = dbt_cloud_group.test_group.id
}
`, groupName)
}
