package dbt_cloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudUserDataSource(t *testing.T) {

	userEmail := "gary.james19@gmail.com"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
data "dbt_cloud_user" "test_user_read" {
    email = "%s"
}
`, userEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dbt-cloud_user.test_user_read", "email", userEmail),
					resource.TestCheckResourceAttrSet("data.dbt-cloud_user.test_user_read", "id"),
				),
			},
		},
	})
}
