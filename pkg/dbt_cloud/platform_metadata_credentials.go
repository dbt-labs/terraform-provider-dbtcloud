package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// PlatformMetadataCredentialsResponse represents the API response for a single credential
type PlatformMetadataCredentialsResponse struct {
	Data   PlatformMetadataCredential `json:"data"`
	Status ResponseStatus             `json:"status"`
}

// PlatformMetadataCredentialsListResponse represents the API response for listing credentials
type PlatformMetadataCredentialsListResponse struct {
	Data   []PlatformMetadataCredential `json:"data"`
	Status ResponseStatus               `json:"status"`
}

// PlatformMetadataCredential represents a platform metadata credential
type PlatformMetadataCredential struct {
	ID                      *int64                           `json:"id,omitempty"`
	AccountID               int64                            `json:"account_id,omitempty"`
	ConnectionID            int64                            `json:"connection_id,omitempty"`
	AdapterVersion          string                           `json:"adapter_version,omitempty"`
	CatalogIngestionEnabled bool                             `json:"catalog_ingestion_enabled"`
	CostOptimizationEnabled bool                             `json:"cost_optimization_enabled"`
	CostInsightsEnabled     bool                             `json:"cost_insights_enabled"`
	Config                  PlatformMetadataCredentialConfig `json:"config"`
	CreatedAt               string                           `json:"created_at,omitempty"`
	UpdatedAt               string                           `json:"updated_at,omitempty"`
}

// PlatformMetadataCredentialConfig represents the adapter-specific configuration
// This uses a flexible structure to support different adapters (Snowflake, Databricks)
type PlatformMetadataCredentialConfig struct {
	// Common/Snowflake fields
	AuthType             string `json:"auth_type,omitempty"`
	User                 string `json:"user,omitempty"`
	Password             string `json:"password,omitempty"`
	PrivateKey           string `json:"private_key,omitempty"`
	PrivateKeyPassphrase string `json:"private_key_passphrase,omitempty"`
	Role                 string `json:"role,omitempty"`
	Warehouse            string `json:"warehouse,omitempty"`

	// Databricks fields
	Token   string `json:"token,omitempty"`
	Catalog string `json:"catalog,omitempty"`
}

// GetPlatformMetadataCredential retrieves a single platform metadata credential by ID
func (c *Client) GetPlatformMetadataCredential(credentialID int64) (*PlatformMetadataCredential, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%d/platform-metadata-credentials/%d/",
			c.HostURL,
			c.AccountID,
			credentialID,
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

	response := PlatformMetadataCredentialsResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// ListPlatformMetadataCredentials retrieves all platform metadata credentials for the account
func (c *Client) ListPlatformMetadataCredentials() ([]PlatformMetadataCredential, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%d/platform-metadata-credentials/",
			c.HostURL,
			c.AccountID,
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

	response := PlatformMetadataCredentialsListResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// CreatePlatformMetadataCredential creates a new platform metadata credential
func (c *Client) CreatePlatformMetadataCredential(
	credential PlatformMetadataCredential,
) (*PlatformMetadataCredential, error) {
	credential.AccountID = int64(c.AccountID)

	requestData, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/platform-metadata-credentials/",
			c.HostURL,
			c.AccountID,
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

	response := PlatformMetadataCredentialsResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// UpdatePlatformMetadataCredential updates an existing platform metadata credential
func (c *Client) UpdatePlatformMetadataCredential(
	credentialID int64,
	credential PlatformMetadataCredential,
) (*PlatformMetadataCredential, error) {
	credential.AccountID = int64(c.AccountID)

	requestData, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf(
			"%s/v3/accounts/%d/platform-metadata-credentials/%d/",
			c.HostURL,
			c.AccountID,
			credentialID,
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

	response := PlatformMetadataCredentialsResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// DeletePlatformMetadataCredential soft-deletes a platform metadata credential
func (c *Client) DeletePlatformMetadataCredential(credentialID int64) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%d/platform-metadata-credentials/%d/",
			c.HostURL,
			c.AccountID,
			credentialID,
		),
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithRetry(req)
	return err
}

// TriggerPlatformMetadataIngestion triggers catalog ingestion for a credential
func (c *Client) TriggerPlatformMetadataIngestion(credentialID int64) error {
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/platform-metadata-credentials/%d/trigger-ingestion/",
			c.HostURL,
			c.AccountID,
			credentialID,
		),
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithRetry(req)
	return err
}
