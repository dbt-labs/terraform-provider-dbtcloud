package group_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDbtCloudGroupsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudGroupsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dbtcloud_groups.test", "groups.#"),
					resource.TestCheckResourceAttrSet("data.dbtcloud_groups.test", "groups.0.id"),
					resource.TestCheckResourceAttrSet("data.dbtcloud_groups.test", "groups.0.name"),
					resource.TestCheckResourceAttrSet("data.dbtcloud_groups.test", "groups.0.state"),
					resource.TestCheckResourceAttrSet("data.dbtcloud_groups.test", "groups.0.scim_managed"),
				),
			},
			{
				Config: testAccDbtCloudGroupsDataSourceConfigWithFilters(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dbtcloud_groups.test_filtered", "groups.#"),
				),
			},
			{
				Config: testAccDbtCloudGroupsDataSourceConfigWithIntegerState(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dbtcloud_groups.test_integer_state", "groups.#"),
				),
			},
		},
	})
}

func testAccDbtCloudGroupsDataSourceConfig() string {
	return fmt.Sprintf(`
data "dbtcloud_groups" "test" {}
`)
}

func testAccDbtCloudGroupsDataSourceConfigWithFilters() string {
	return fmt.Sprintf(`
data "dbtcloud_groups" "test_filtered" {
  state = "active"
  name_contains = "member"
}
`)
}

func testAccDbtCloudGroupsDataSourceConfigWithIntegerState() string {
	return fmt.Sprintf(`
data "dbtcloud_groups" "test_integer_state" {
  state = "1"
}
`)
}
