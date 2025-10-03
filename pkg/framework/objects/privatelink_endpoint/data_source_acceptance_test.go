package privatelink_endpoint_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
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

func TestPrivatelinkEndpointDataSource_PaginationTwoPages(t *testing.T) {
	originalTFAcc := os.Getenv("TF_ACC")
	os.Setenv("TF_ACC", "1")
	defer func() {
		if originalTFAcc == "" {
			os.Unsetenv("TF_ACC")
		} else {
			os.Setenv("TF_ACC", originalTFAcc)
		}
	}()

	// Create 100 endpoints for page 1
	page1Endpoints := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		page1Endpoints[i] = map[string]interface{}{
			"id":                    fmt.Sprintf("ple_%d", i),
			"account_id":            123,
			"name":                  fmt.Sprintf("Endpoint %d", i),
			"type":                  "snowflake",
			"private_link_endpoint": fmt.Sprintf("vpce-%d.snowflakecomputing.com", i),
			"cidr_range":            "10.0.0.0/8",
			"state":                 1,
		}
	}

	// Create 50 endpoints for page 2, including our target endpoint
	page2Endpoints := make([]map[string]interface{}, 50)
	for i := 0; i < 49; i++ {
		page2Endpoints[i] = map[string]interface{}{
			"id":                    fmt.Sprintf("ple_%d", i+100),
			"account_id":            123,
			"name":                  fmt.Sprintf("Endpoint %d", i+100),
			"type":                  "snowflake",
			"private_link_endpoint": fmt.Sprintf("vpce-%d.snowflakecomputing.com", i+100),
			"cidr_range":            "10.0.0.0/8",
			"state":                 1,
		}
	}

	// Set our target endpoint to be the last one on page 2
	targetEndpoint := map[string]interface{}{
		"id":                    "ple_target",
		"account_id":            123,
		"name":                  "Target Endpoint",
		"type":                  "redshift",
		"private_link_endpoint": "vpce-target.redshift.amazonaws.com",
		"cidr_range":            "172.16.0.0/12",
		"state":                 1,
	}
	page2Endpoints[49] = targetEndpoint

	handlers := map[string]testhelpers.MockEndpointHandler{
		"GET /v3/accounts/123/private-link-endpoints/": func(r *http.Request) (int, interface{}, error) {
			offset := r.URL.Query().Get("offset")

			if offset == "100" {
				// Page 2
				response := map[string]interface{}{
					"data": page2Endpoints,
					"extra": map[string]interface{}{
						"pagination": map[string]interface{}{
							"count":       len(page2Endpoints),
							"total_count": 150, // 100 + 50 total endpoints
						},
					},
				}
				return http.StatusOK, response, nil
			} else {
				// Page 1 (offset == "" or offset == "0")
				response := map[string]interface{}{
					"data": page1Endpoints,
					"extra": map[string]interface{}{
						"pagination": map[string]interface{}{
							"count":       len(page1Endpoints),
							"total_count": 150, // 100 + 50 total endpoints
						},
					},
				}
				return http.StatusOK, response, nil
			}
		},
	}

	mockServer := testhelpers.SetupMockServer(t, handlers)
	defer mockServer.Close()

	// Test finding endpoint by name on second page
	configByName := fmt.Sprintf(`
		provider "dbtcloud" {
			host_url   = "%s"
			token      = "test-token"
			account_id = 123
		}

		data "dbtcloud_privatelink_endpoint" "test" {
			name = "Target Endpoint"
		}
	`, mockServer.URL)

	// Test finding endpoint by URL on second page
	configByURL := fmt.Sprintf(`
		provider "dbtcloud" {
			host_url   = "%s"
			token      = "test-token"
			account_id = 123
		}

		data "dbtcloud_privatelink_endpoint" "test" {
			private_link_endpoint_url = "vpce-target.redshift.amazonaws.com"
		}
	`, mockServer.URL)

	// Test finding endpoint by both name and URL on second page
	configByBoth := fmt.Sprintf(`
		provider "dbtcloud" {
			host_url   = "%s"
			token      = "test-token"
			account_id = 123
		}

		data "dbtcloud_privatelink_endpoint" "test" {
			name = "Target Endpoint"
			private_link_endpoint_url = "vpce-target.redshift.amazonaws.com"
		}
	`, mockServer.URL)

	// Test finding endpoint on first page
	configFirstPage := fmt.Sprintf(`
		provider "dbtcloud" {
			host_url   = "%s"
			token      = "test-token"
			account_id = 123
		}

		data "dbtcloud_privatelink_endpoint" "test" {
			name = "Endpoint 50"
		}
	`, mockServer.URL)

	checkTargetEndpoint := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoint.test", "id", "ple_target"),
		resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoint.test", "name", "Target Endpoint"),
		resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoint.test", "type", "redshift"),
		resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoint.test", "private_link_endpoint_url", "vpce-target.redshift.amazonaws.com"),
		resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoint.test", "cidr_range", "172.16.0.0/12"),
	)

	checkFirstPageEndpoint := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoint.test", "id", "ple_50"),
		resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoint.test", "name", "Endpoint 50"),
		resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoint.test", "type", "snowflake"),
		resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoint.test", "private_link_endpoint_url", "vpce-50.snowflakecomputing.com"),
		resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoint.test", "cidr_range", "10.0.0.0/8"),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configByName,
				Check:  checkTargetEndpoint,
			},
			{
				Config: configByURL,
				Check:  checkTargetEndpoint,
			},
			{
				Config: configByBoth,
				Check:  checkTargetEndpoint,
			},
			{
				Config: configFirstPage,
				Check:  checkFirstPageEndpoint,
			},
		},
	})
}
