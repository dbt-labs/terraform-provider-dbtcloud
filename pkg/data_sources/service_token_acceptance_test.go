package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudServiceTokenDataSource(t *testing.T) {

	serviceTokenName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := serviceToken(serviceTokenName)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbt_cloud_service_token.test_service_token_read", "name", serviceTokenName),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_service_token.test_service_token_read", "uid"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_service_token.test_service_token_read", "service_token_id"),
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
resource "dbt_cloud_service_token" "test_service_token" {
    name = "%s"
}

data "dbt_cloud_service_token" "test_service_token_read" {
    service_token_id = dbt_cloud_service_token.test_service_token.id
}
`, serviceTokenName)
}
