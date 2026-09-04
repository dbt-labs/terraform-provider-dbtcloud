package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type OpenAIIntegration struct {
	ID                  *int64  `json:"id,omitempty"`
	AccountID           int64   `json:"account_id,omitempty"`
	KeyType             string  `json:"key_type"`
	KeyValue            *string `json:"key_value,omitempty"`
	AzureEndpoint       *string `json:"azure_endpoint,omitempty"`
	AzureDeploymentName *string `json:"azure_deployment_name,omitempty"`
	AzureAPIVersion     *string `json:"azure_api_version,omitempty"`
	CreatedAt           *string `json:"created_at,omitempty"`
	UpdatedAt           *string `json:"updated_at,omitempty"`
}

type OpenAIIntegrationResponse struct {
	Data   OpenAIIntegration `json:"data"`
	Status ResponseStatus    `json:"status"`
}

type OpenAIIntegrationListResponse struct {
	Data   []OpenAIIntegration `json:"data"`
	Status ResponseStatus      `json:"status"`
}

func (c *Client) GetOpenAIIntegration(integrationID int64) (*OpenAIIntegration, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/integrations/open-ai/%d/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			integrationID,
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

	resp := OpenAIIntegrationResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) CreateOpenAIIntegration(integration OpenAIIntegration) (*OpenAIIntegration, error) {
	// account_id is passed via the URL path — the API rejects it in the body.
	integration.AccountID = 0

	payload, err := json.Marshal(integration)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/integrations/open-ai/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
		),
		strings.NewReader(string(payload)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	resp := OpenAIIntegrationResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) UpdateOpenAIIntegration(integrationID int64, integration OpenAIIntegration) (*OpenAIIntegration, error) {
	// account_id is passed via the URL path; zero it out to avoid API rejection.
	integration.AccountID = 0

	payload, err := json.Marshal(integration)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf(
			"%s/v3/accounts/%s/integrations/open-ai/%d/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			integrationID,
		),
		strings.NewReader(string(payload)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	resp := OpenAIIntegrationResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) DeleteOpenAIIntegration(integrationID int64) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%s/integrations/open-ai/%d/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			integrationID,
		),
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithRetry(req)
	return err
}

// GetAllOpenAIIntegrations lists the OpenAI integrations of the account. The API allows
// one per account, so this returns at most one entry.
func (c *Client) GetAllOpenAIIntegrations() ([]OpenAIIntegration, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/integrations/open-ai/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
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

	resp := OpenAIIntegrationListResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}
