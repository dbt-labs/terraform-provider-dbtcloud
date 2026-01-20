package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ConnectionCatalogConfigResponse represents the API response for catalog config
type ConnectionCatalogConfigResponse struct {
	Data   ConnectionCatalogConfig `json:"data"`
	Status ResponseStatus          `json:"status"`
}

// ConnectionCatalogConfig represents the catalog configuration for a connection
// Note: omitempty is used so that nil slices are omitted from the JSON payload
// This is important because the API expects fields to be absent rather than null
type ConnectionCatalogConfig struct {
	ConnectionID  int64    `json:"connection_id,omitempty"`
	DatabaseAllow []string `json:"database_allow,omitempty"`
	DatabaseDeny  []string `json:"database_deny,omitempty"`
	SchemaAllow   []string `json:"schema_allow,omitempty"`
	SchemaDeny    []string `json:"schema_deny,omitempty"`
	TableAllow    []string `json:"table_allow,omitempty"`
	TableDeny     []string `json:"table_deny,omitempty"`
	ViewAllow     []string `json:"view_allow,omitempty"`
	ViewDeny      []string `json:"view_deny,omitempty"`
}

// GetConnectionCatalogConfig retrieves the catalog configuration for a connection
func (c *Client) GetConnectionCatalogConfig(connectionID int64) (*ConnectionCatalogConfig, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%d/connections/%d/catalog-configs/",
			c.HostURL,
			c.AccountID,
			connectionID,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	response := ConnectionCatalogConfigResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// UpdateConnectionCatalogConfig updates the catalog configuration for a connection
// This is used for both Create and Update operations since there's no POST endpoint
func (c *Client) UpdateConnectionCatalogConfig(
	connectionID int64,
	config ConnectionCatalogConfig,
) (*ConnectionCatalogConfig, error) {
	requestData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf(
			"%s/v3/accounts/%d/connections/%d/catalog-configs/",
			c.HostURL,
			c.AccountID,
			connectionID,
		),
		strings.NewReader(string(requestData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	response := ConnectionCatalogConfigResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// DeleteConnectionCatalogConfig "deletes" the catalog config by setting all fields to empty arrays
// There's no DELETE endpoint, so we PATCH with empty arrays to clear the configuration
func (c *Client) DeleteConnectionCatalogConfig(connectionID int64) error {
	// Send empty arrays for all filter fields to clear the configuration
	payload := `{"database_allow":[],"database_deny":[],"schema_allow":[],"schema_deny":[],"table_allow":[],"table_deny":[],"view_allow":[],"view_deny":[]}`

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf(
			"%s/v3/accounts/%d/connections/%d/catalog-configs/",
			c.HostURL,
			c.AccountID,
			connectionID,
		),
		strings.NewReader(payload),
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithRetry(req)
	return err
}
