package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// AnalyticsCredentialResponse represents the API response for a single analytics credential
type AnalyticsCredentialResponse struct {
	Data   AnalyticsCredential `json:"data"`
	Status ResponseStatus      `json:"status"`
}

// AnalyticsCredential represents an analytics credential as returned by the API
type AnalyticsCredential struct {
	ID             *int64                    `json:"id,omitempty"`
	AccountID      int64                     `json:"account_id,omitempty"`
	ConnectionID   int64                     `json:"connection_id,omitempty"`
	AdapterVersion string                    `json:"adapter_version,omitempty"`
	Config         AnalyticsCredentialConfig `json:"config"`
	CreatedAt      string                    `json:"created_at,omitempty"`
	UpdatedAt      string                    `json:"updated_at,omitempty"`
}

// AnalyticsCredentialConfig is the adapter-specific configuration. Fields are only
// sent when set; the Redshift and BigQuery fields are added with their adapter blocks.
type AnalyticsCredentialConfig struct {
	// Snowflake fields
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

	// Redshift fields (the password is shared with Snowflake)
	Username string `json:"username,omitempty"`
}

// AnalyticsCredentialCreateRequest is the create body; the adapter comes from the connection.
type AnalyticsCredentialCreateRequest struct {
	AccountID    int64                     `json:"account_id"`
	ConnectionID int64                     `json:"connection_id"`
	Config       AnalyticsCredentialConfig `json:"config"`
}

// AnalyticsCredentialUpdateRequest is the update body; connection and adapter cannot change.
type AnalyticsCredentialUpdateRequest struct {
	AccountID int64                     `json:"account_id"`
	Config    AnalyticsCredentialConfig `json:"config"`
}

func (c *Client) analyticsCredentialsURL() string {
	return fmt.Sprintf("%s/v3/accounts/%d/analytics-credentials/", c.HostURL, c.AccountID)
}

// GetAnalyticsCredential retrieves a single analytics credential by ID
func (c *Client) GetAnalyticsCredential(credentialID int64) (*AnalyticsCredential, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s%d/", c.analyticsCredentialsURL(), credentialID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	response := AnalyticsCredentialResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// ListAnalyticsCredentials retrieves every analytics credential of the account,
// optionally filtered by connection and adapter version.
func (c *Client) ListAnalyticsCredentials(
	connectionID *int64,
	adapterVersion string,
) ([]AnalyticsCredential, error) {
	query := url.Values{}
	if connectionID != nil {
		query.Set("connection_id", strconv.FormatInt(*connectionID, 10))
	}
	if adapterVersion != "" {
		query.Set("adapter_version", adapterVersion)
	}

	listURL := c.analyticsCredentialsURL()
	if len(query) > 0 {
		listURL += "?" + query.Encode()
	}

	rawItems, err := c.GetRawData(listURL)
	if err != nil {
		return nil, err
	}

	credentials := make([]AnalyticsCredential, 0, len(rawItems))
	for _, raw := range rawItems {
		credential := AnalyticsCredential{}
		if err := json.Unmarshal(raw, &credential); err != nil {
			return nil, err
		}
		credentials = append(credentials, credential)
	}

	return credentials, nil
}

// CreateAnalyticsCredential creates an analytics credential on a connection
func (c *Client) CreateAnalyticsCredential(
	connectionID int64,
	config AnalyticsCredentialConfig,
) (*AnalyticsCredential, error) {
	requestData, err := json.Marshal(AnalyticsCredentialCreateRequest{
		AccountID:    int64(c.AccountID),
		ConnectionID: connectionID,
		Config:       config,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		c.analyticsCredentialsURL(),
		strings.NewReader(string(requestData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	response := AnalyticsCredentialResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// UpdateAnalyticsCredential updates the config of an analytics credential
func (c *Client) UpdateAnalyticsCredential(
	credentialID int64,
	config AnalyticsCredentialConfig,
) (*AnalyticsCredential, error) {
	requestData, err := json.Marshal(AnalyticsCredentialUpdateRequest{
		AccountID: int64(c.AccountID),
		Config:    config,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf("%s%d/", c.analyticsCredentialsURL(), credentialID),
		strings.NewReader(string(requestData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	response := AnalyticsCredentialResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// DeleteAnalyticsCredential deletes an analytics credential
func (c *Client) DeleteAnalyticsCredential(credentialID int64) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s%d/", c.analyticsCredentialsURL(), credentialID),
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithRetry(req)
	return err
}
