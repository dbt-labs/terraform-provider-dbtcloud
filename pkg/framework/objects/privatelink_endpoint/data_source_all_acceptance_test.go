package privatelink_endpoint_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
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
		if strings.Contains(r.URL.Path, "privatelink-endpoints") {
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

	client := &dbt_cloud.Client{
		HostURL:    mockServer.URL,
		HTTPClient: &http.Client{},
		AccountID:  1,
	}

	privatelink_endpoints, err := client.GetAllPrivatelinkEndpoints()

	assert.NoError(t, err)
	assert.Equal(t, 101, len(privatelink_endpoints))
	assert.Equal(t, "privatelink_endpoint100", privatelink_endpoints[100].ID)
}
