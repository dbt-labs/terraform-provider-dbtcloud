package user_test

import (
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudUsersDataSource(t *testing.T) {

	var userEmail string
	if acctest_helper.IsDbtCloudPR() {
		userEmail = "d" + "ev@" + "db" + "tla" + "bs.c" + "om"
	} else {
		userEmail = "beno" + "it" + ".per" + "igaud" + "@" + "fisht" + "ownanalytics" + "." + "com"
	}

	_ = userEmail
	config := users()

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.dbtcloud_users.all", "users.0.email"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_users.all", "users.0.id"),
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

func users() string {
	return `
data "dbtcloud_users" "all" {
}
`
}
