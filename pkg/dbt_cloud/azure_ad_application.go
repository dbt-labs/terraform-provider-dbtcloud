package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type AzureADApplication struct {
	ID                               *int64  `json:"id,omitempty"`
	AccountID                        int64   `json:"account_id,omitempty"`
	OrganizationName                 string  `json:"organization_name"`
	ClientID                         *string `json:"client_id,omitempty"`
	ClientSecret                     *string `json:"client_secret,omitempty"`
	TenantID                         *string `json:"tenant_id,omitempty"`
	AzureServiceAuthenticationMethod string  `json:"azure_service_authentication_method,omitempty"`
	OAuthRedirectURIDomain           *string `json:"oauth_redirect_uri_domain,omitempty"`
	CreatedAt                        *string `json:"created_at,omitempty"`
	UpdatedAt                        *string `json:"updated_at,omitempty"`
}

type AzureADApplicationResponse struct {
	Data   AzureADApplication `json:"data"`
	Status ResponseStatus     `json:"status"`
}

type AzureADApplicationListResponse struct {
	Data   []AzureADApplication `json:"data"`
	Status ResponseStatus       `json:"status"`
}

// GetAzureADApplicationForAccount fetches the single Azure AD application for
// this account via the list endpoint. Returns nil if none exists yet.
func (c *Client) GetAzureADApplicationForAccount() (*AzureADApplication, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/azure-ad-applications/",
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

	listResp := AzureADApplicationListResponse{}
	if err = json.Unmarshal(body, &listResp); err != nil {
		return nil, err
	}

	if len(listResp.Data) == 0 {
		return nil, nil
	}
	return &listResp.Data[0], nil
}

func (c *Client) GetAzureADApplication(applicationID int64) (*AzureADApplication, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/azure-ad-applications/%d/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			applicationID,
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

	resp := AzureADApplicationResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) CreateAzureADApplication(app AzureADApplication) (*AzureADApplication, error) {
	payload, err := json.Marshal(app)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/azure-ad-applications/",
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

	resp := AzureADApplicationResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) UpdateAzureADApplication(applicationID int64, app AzureADApplication) (*AzureADApplication, error) {
	// account_id is passed via the URL path; zero it out to avoid API rejection.
	app.AccountID = 0

	payload, err := json.Marshal(app)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/azure-ad-applications/%d/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			applicationID,
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

	resp := AzureADApplicationResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) DeleteAzureADApplication(applicationID int64) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%s/azure-ad-applications/%d/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			applicationID,
		),
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithRetry(req)
	return err
}
