package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudServiceTokenDataSource(t *testing.T) {

	serviceTokenName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := serviceToken(serviceTokenName)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(
			"data.dbtcloud_service_token.test_service_token_read",
			"name",
			serviceTokenName,
		),
		resource.TestCheckResourceAttrSet(
			"data.dbtcloud_service_token.test_service_token_read",
			"uid",
		),
		resource.TestCheckResourceAttrSet(
			"data.dbtcloud_service_token.test_service_token_read",
			"service_token_id",
		),
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

func serviceToken(serviceTokenName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_service_token" "test_service_token" {
    name = "%s"
}

data "dbtcloud_service_token" "test_service_token_read" {
    service_token_id = dbtcloud_service_token.test_service_token.id
}
`, serviceTokenName)
}
