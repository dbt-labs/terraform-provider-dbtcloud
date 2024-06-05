package data_sources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudPrivatelinkEndpointDataSource(t *testing.T) {

	// we only test this explicitly as we can't create resources and need to read from existing ones
	if os.Getenv("DBT_ACCEPTANCE_TEST_PRIVATE_LINK") == "" {
		t.Skip("Skipping acceptance tests as DBT_ACCEPTANCE_TEST_PRIVATE_LINK is not set")
	}

	endpointName := os.Getenv("DBT_ACCEPTANCE_TEST_PRIVATE_LINK_NAME")
	endpointURL := os.Getenv("DBT_ACCEPTANCE_TEST_PRIVATE_LINK_URL")

	// different configurations whether we provide the endpoint name and/or url
	config := privatelinkEndpoint(endpointName, endpointURL)
	configNoURL := privatelinkEndpoint(endpointName, "")

	configNoName := privatelinkEndpoint("", endpointURL)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(
			"data.dbtcloud_privatelink_endpoint.test",
			"name",
			endpointName,
		),
		resource.TestCheckResourceAttr(
			"data.dbtcloud_privatelink_endpoint.test",
			"private_link_endpoint_url",
			endpointURL,
		),
		resource.TestCheckResourceAttrSet("data.dbtcloud_privatelink_endpoint.test", "id"),
		resource.TestCheckResourceAttrSet(
			"data.dbtcloud_privatelink_endpoint.test",
			"cidr_range",
		),
		resource.TestCheckResourceAttrSet("data.dbtcloud_privatelink_endpoint.test", "type"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_privatelink_endpoint.test", "state"),
	)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  check,
			},
			{
				Config: configNoURL,
				Check:  check,
			},
			{
				Config: configNoName,
				Check:  check,
			},
		},
	})
}

func privatelinkEndpoint(endpointName, endpointURL string) string {
	return fmt.Sprintf(`
	data "dbtcloud_privatelink_endpoint" "test" {
		name = "%s"
		private_link_endpoint_url = "%s"
	  }
    `, endpointName, endpointURL)
}
