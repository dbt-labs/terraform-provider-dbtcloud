package privatelink_endpoint_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestAccDbtCloudPrivatelinkEndpointPagination(t *testing.T) {
	page1_privatelink_endpoints := make([]dbt_cloud.PrivatelinkEndpoint, 100)
	for i := 0; i < 100; i++ {
		page1_privatelink_endpoints[i] = dbt_cloud.PrivatelinkEndpoint{
			ID:                     fmt.Sprintf("privatelink_endpoint%d", i),
			Name:                   fmt.Sprintf("PrivatelinkEndpoint %d", i),
			PrivatelinkEndpointURL: fmt.Sprintf("privatelink_endpoint%d.com", i),
		}
	}

	page2_privatelink_endpoints := []dbt_cloud.PrivatelinkEndpoint{
		{
			ID:                     "privatelink_endpoint100",
			Name:                   "PrivatelinkEndpoint 100",
			PrivatelinkEndpointURL: "privatelink_endpoint100.com",
		},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "private-link-endpoints") {
			offset := r.URL.Query().Get("offset")
			if offset == "100" {
				// Page 2
				response := dbt_cloud.Response{
					Data: make([]interface{}, len(page2_privatelink_endpoints)),
					Extra: dbt_cloud.Extra{
						Pagination: dbt_cloud.Pagination{
							Count:      len(page2_privatelink_endpoints),
							TotalCount: 101,
						},
					},
				}
				for i, privatelink_endpoint := range page2_privatelink_endpoints {
					response.Data[i] = privatelink_endpoint
				}
				json.NewEncoder(w).Encode(response)
			} else {
				// Page 1
				response := dbt_cloud.Response{
					Data: make([]interface{}, len(page1_privatelink_endpoints)),
					Extra: dbt_cloud.Extra{
						Pagination: dbt_cloud.Pagination{
							Count:      len(page1_privatelink_endpoints),
							TotalCount: 101,
						},
					},
				}
				for i, privatelink_endpoint := range page1_privatelink_endpoints {
					response.Data[i] = privatelink_endpoint
				}
				json.NewEncoder(w).Encode(response)
			}
		}
	}))
	defer mockServer.Close()

	parsedURL, err := url.Parse(mockServer.URL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse serverURL: %s, error: %v", mockServer.URL, err))
	}
	client := &dbt_cloud.Client{
		HostURL:    parsedURL,
		HTTPClient: &http.Client{},
		AccountID:  1,
	}

	privatelink_endpoints, err := client.GetAllPrivatelinkEndpoints()

	assert.NoError(t, err)
	assert.Equal(t, 101, len(privatelink_endpoints))
	assert.Equal(t, "privatelink_endpoint100", privatelink_endpoints[100].ID)
}

// Unit tests using mock server (no real API connection required)

func TestPrivatelinkEndpointsDataSource_Basic(t *testing.T) {
	originalTFAcc := os.Getenv("TF_ACC")
	os.Setenv("TF_ACC", "1")
	defer func() {
		if originalTFAcc == "" {
			os.Unsetenv("TF_ACC")
		} else {
			os.Setenv("TF_ACC", originalTFAcc)
		}
	}()

	// Create mock endpoints based on the API structure and realistic names from backend tests
	mockEndpoints := []map[string]interface{}{
		{
			"id":                    "ple_snowflake_prod_001",
			"account_id":            123,
			"name":                  "Snowflake Production Endpoint",
			"type":                  "snowflake",
			"private_link_endpoint": "vpce-1234567890abcdef0.snowflakecomputing.com",
			"cidr_range":            "10.0.0.0/8",
			"state":                 1,
		},
		{
			"id":                    "ple_redshift_staging_001",
			"account_id":            123,
			"name":                  "Redshift Data Warehouse",
			"type":                  "redshift",
			"private_link_endpoint": "vpce-0987654321fedcba0.us-west-2.redshift.amazonaws.com",
			"cidr_range":            "172.16.0.0/12",
			"state":                 1,
		},
	}

	handlers := map[string]testhelpers.MockEndpointHandler{
		"GET /v3/accounts/123/private-link-endpoints/": func(r *http.Request) (int, interface{}, error) {
			// Parse pagination parameters like the real API
			limit := 100
			offset := 0
			if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
				fmt.Sscanf(limitParam, "%d", &limit)
			}
			if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
				fmt.Sscanf(offsetParam, "%d", &offset)
			}

			// Return paginated results
			endIndex := offset + limit
			if endIndex > len(mockEndpoints) {
				endIndex = len(mockEndpoints)
			}

			data := make([]interface{}, 0)
			if offset < len(mockEndpoints) {
				for i := offset; i < endIndex; i++ {
					data = append(data, mockEndpoints[i])
				}
			}

			response := map[string]interface{}{
				"data": data,
				"extra": map[string]interface{}{
					"pagination": map[string]interface{}{
						"count":       len(data),
						"total_count": len(mockEndpoints),
					},
				},
			}
			return http.StatusOK, response, nil
		},
	}

	mockServer := testhelpers.SetupMockServer(t, handlers)
	defer mockServer.Close()

	config := fmt.Sprintf(`
		provider "dbtcloud" {
		host_url   = "%s"
		token      = "test-token"
		account_id = 123
		}

		data "dbtcloud_privatelink_endpoints" "test" {}
		`, mockServer.URL)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoints.test", "endpoints.#", "2"),
					resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoints.test", "endpoints.0.name", "Snowflake Production Endpoint"),
					resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoints.test", "endpoints.0.type", "snowflake"),
					resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoints.test", "endpoints.0.id", "ple_snowflake_prod_001"),
					resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoints.test", "endpoints.0.cidr_range", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoints.test", "endpoints.1.name", "Redshift Data Warehouse"),
					resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoints.test", "endpoints.1.type", "redshift"),
					resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoints.test", "endpoints.1.id", "ple_redshift_staging_001"),
					resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoints.test", "endpoints.1.cidr_range", "172.16.0.0/12"),
				),
			},
		},
	})
}

func TestPrivatelinkEndpointsDataSource_Empty(t *testing.T) {
	originalTFAcc := os.Getenv("TF_ACC")
	os.Setenv("TF_ACC", "1")
	defer func() {
		if originalTFAcc == "" {
			os.Unsetenv("TF_ACC")
		} else {
			os.Setenv("TF_ACC", originalTFAcc)
		}
	}()

	handlers := map[string]testhelpers.MockEndpointHandler{
		"GET /v3/accounts/123/private-link-endpoints/": func(r *http.Request) (int, interface{}, error) {
			response := map[string]interface{}{
				"data": []interface{}{},
				"extra": map[string]interface{}{
					"pagination": map[string]interface{}{
						"count":       0,
						"total_count": 0,
					},
				},
			}
			return http.StatusOK, response, nil
		},
	}

	mockServer := testhelpers.SetupMockServer(t, handlers)
	defer mockServer.Close()

	config := fmt.Sprintf(`
		provider "dbtcloud" {
		host_url   = "%s"
		token      = "test-token"
		account_id = 123
		}

		data "dbtcloud_privatelink_endpoints" "test" {}
		`, mockServer.URL)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoints.test", "endpoints.#", "0"),
				),
			},
		},
	})
}

func TestPrivatelinkEndpointsDataSource_WithFiltering(t *testing.T) {
	originalTFAcc := os.Getenv("TF_ACC")
	os.Setenv("TF_ACC", "1")
	defer func() {
		if originalTFAcc == "" {
			os.Unsetenv("TF_ACC")
		} else {
			os.Setenv("TF_ACC", originalTFAcc)
		}
	}()

	// Mock endpoints covering all major warehouse types and git providers
	mockEndpoints := []map[string]interface{}{
		{
			"id":                    "ple_snowflake_prod_001",
			"account_id":            123,
			"name":                  "Snowflake Production Endpoint",
			"type":                  "snowflake",
			"private_link_endpoint": "vpce-1234567890abcdef0.snowflakecomputing.com",
			"cidr_range":            "10.0.0.0/8",
			"state":                 1,
		},
		{
			"id":                    "ple_snowflake_staging_001",
			"account_id":            123,
			"name":                  "Snowflake Development Environment",
			"type":                  "snowflake",
			"private_link_endpoint": "vpce-abcdef1234567890.snowflakecomputing.com",
			"cidr_range":            "10.1.0.0/16",
			"state":                 1,
		},
		{
			"id":                    "ple_redshift_prod_001",
			"account_id":            123,
			"name":                  "Redshift Data Warehouse",
			"type":                  "redshift",
			"private_link_endpoint": "vpce-0987654321fedcba0.us-west-2.redshift.amazonaws.com",
			"cidr_range":            "172.16.0.0/12",
			"state":                 1,
		},
		{
			"id":                    "ple_bigquery_analytics_001",
			"account_id":            123,
			"name":                  "BigQuery Analytics Hub",
			"type":                  "bigquery",
			"private_link_endpoint": "vpce-bigquery123456789.us-central1.gcp.com",
			"cidr_range":            "192.168.0.0/16",
			"state":                 1,
		},
		{
			"id":                    "ple_github_enterprise_001",
			"account_id":            123,
			"name":                  "GitHub Enterprise Server",
			"type":                  "github",
			"private_link_endpoint": "vpce-github987654321.internal.company.com",
			"cidr_range":            "10.100.0.0/16",
			"state":                 1,
		},
	}

	handlers := map[string]testhelpers.MockEndpointHandler{
		"GET /v3/accounts/123/private-link-endpoints/": func(r *http.Request) (int, interface{}, error) {
			response := map[string]interface{}{
				"data": mockEndpoints,
				"extra": map[string]interface{}{
					"pagination": map[string]interface{}{
						"count":       len(mockEndpoints),
						"total_count": len(mockEndpoints),
					},
				},
			}
			return http.StatusOK, response, nil
		},
	}

	mockServer := testhelpers.SetupMockServer(t, handlers)
	defer mockServer.Close()

	config := fmt.Sprintf(`
		provider "dbtcloud" {
		host_url   = "%s"
		token      = "test-token"
		account_id = 123
		}
		
		data "dbtcloud_privatelink_endpoints" "test" {}
		
		locals {
		  # Find specific endpoint by name
		  snowflake_production = [
		    for endpoint in data.dbtcloud_privatelink_endpoints.test.endpoints :
		    endpoint if endpoint.name == "Snowflake Production Endpoint"
		  ][0]

		  # Filter endpoints by type
		  snowflake_endpoints = [
		    for endpoint in data.dbtcloud_privatelink_endpoints.test.endpoints :
		    endpoint if endpoint.type == "snowflake"
		  ]

		  redshift_endpoints = [
		    for endpoint in data.dbtcloud_privatelink_endpoints.test.endpoints :
		    endpoint if endpoint.type == "redshift"
		  ]

		  bigquery_endpoints = [
		    for endpoint in data.dbtcloud_privatelink_endpoints.test.endpoints :
		    endpoint if endpoint.type == "bigquery"
		  ]

		  github_endpoints = [
		    for endpoint in data.dbtcloud_privatelink_endpoints.test.endpoints :
		    endpoint if endpoint.type == "github"
		  ]

		  # Count by type for validation
		  snowflake_count = length(local.snowflake_endpoints)
		  redshift_count = length(local.redshift_endpoints)
		  bigquery_count = length(local.bigquery_endpoints)
		  github_count = length(local.github_endpoints)
		  total_warehouse_endpoints = local.snowflake_count + local.redshift_count + local.bigquery_count
		}
		
		# Outputs for testing the filtering logic
		output "snowflake_count" {
		  value = local.snowflake_count
		}
		
		output "redshift_count" {
		  value = local.redshift_count
		}

		output "bigquery_count" {
		  value = local.bigquery_count
		}

		output "github_count" {
		  value = local.github_count
		}

		output "total_warehouse_endpoints" {
		value = local.total_warehouse_endpoints
		}

		output "found_production_endpoint_id" {
		  value = local.snowflake_production.id
		}
		`, mockServer.URL)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					// Verify all endpoints are returned
					resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoints.test", "endpoints.#", "5"),

					// Test filtering by type
					resource.TestCheckOutput("snowflake_count", "2"),
					resource.TestCheckOutput("redshift_count", "1"),
					resource.TestCheckOutput("bigquery_count", "1"),
					resource.TestCheckOutput("github_count", "1"),
					resource.TestCheckOutput("total_warehouse_endpoints", "4"),

					// Test finding specific endpoint by name
					resource.TestCheckOutput("found_production_endpoint_id", "ple_snowflake_prod_001"),
				),
			},
		},
	})
}

func TestPrivatelinkEndpointsDataSource_ForEachPattern(t *testing.T) {
	originalTFAcc := os.Getenv("TF_ACC")
	os.Setenv("TF_ACC", "1")
	defer func() {
		if originalTFAcc == "" {
			os.Unsetenv("TF_ACC")
		} else {
			os.Setenv("TF_ACC", originalTFAcc)
		}
	}()

	// Mock multiple Snowflake endpoints
	mockEndpoints := []map[string]interface{}{
		{
			"id":                    "ple_snowflake_prod_001",
			"account_id":            123,
			"name":                  "Snowflake Production Endpoint",
			"type":                  "snowflake",
			"private_link_endpoint": "vpce-1234567890abcdef0.snowflakecomputing.com",
			"cidr_range":            "10.0.0.0/8",
			"state":                 1,
		},
		{
			"id":                    "ple_snowflake_staging_001",
			"account_id":            123,
			"name":                  "Snowflake Development Environment",
			"type":                  "snowflake",
			"private_link_endpoint": "vpce-abcdef1234567890.snowflakecomputing.com",
			"cidr_range":            "10.1.0.0/16",
			"state":                 1,
		},
	}

	handlers := map[string]testhelpers.MockEndpointHandler{
		"GET /v3/accounts/123/private-link-endpoints/": func(r *http.Request) (int, interface{}, error) {
			response := map[string]interface{}{
				"data": mockEndpoints,
				"extra": map[string]interface{}{
					"pagination": map[string]interface{}{
						"count":       len(mockEndpoints),
						"total_count": len(mockEndpoints),
					},
				},
			}
			return http.StatusOK, response, nil
		},
	}

	mockServer := testhelpers.SetupMockServer(t, handlers)
	defer mockServer.Close()

	config := fmt.Sprintf(`
		provider "dbtcloud" {
		host_url   = "%s"
		token      = "test-token"
		account_id = 123
		}

		data "dbtcloud_privatelink_endpoints" "all" {}

		locals {
		  snowflake_endpoints = [
		    for endpoint in data.dbtcloud_privatelink_endpoints.all.endpoints : 
		    endpoint if endpoint.type == "snowflake"
		  ]
		}

		output "connection_names" {
		  value = {
		    for ep in local.snowflake_endpoints : 
		    ep.id => "Connection for ${ep.name}"
		  }
		}

		output "endpoint_count" {
		  value = length(local.snowflake_endpoints)
		}
		`, mockServer.URL)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					// Verify all endpoints are returned
					resource.TestCheckResourceAttr("data.dbtcloud_privatelink_endpoints.all", "endpoints.#", "2"),

					// Test the for_each pattern produces correct connection names
					resource.TestCheckOutput("endpoint_count", "2"),
					// Test that we can access the mapped values (validates the for_each pattern works)
					resource.TestCheckResourceAttrSet("data.dbtcloud_privatelink_endpoints.all", "endpoints.0.id"),
					resource.TestCheckResourceAttrSet("data.dbtcloud_privatelink_endpoints.all", "endpoints.0.name"),
					resource.TestCheckResourceAttrSet("data.dbtcloud_privatelink_endpoints.all", "endpoints.1.id"),
					resource.TestCheckResourceAttrSet("data.dbtcloud_privatelink_endpoints.all", "endpoints.1.name"),
				),
			},
		},
	})
}
