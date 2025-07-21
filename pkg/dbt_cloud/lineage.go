package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type LineageIntegrationConfig struct {
	Host      string `json:"host,omitempty"`
	SiteID    string `json:"site_id,omitempty"`
	TokenName string `json:"token_name,omitempty"`
	Token     string `json:"token,omitempty"`
}

type LineageIntegration struct {
	ID        *int64                   `json:"id,omitempty"`
	AccountID int64                    `json:"account_id,omitempty"`
	ProjectID int64                    `json:"project_id,omitempty"`
	Name      string                   `json:"name,omitempty"`
	Config    LineageIntegrationConfig `json:"config"`
}

type LineageIntegrationResponse struct {
	Data   LineageIntegration `json:"data"`
	Status ResponseStatus     `json:"status"`
}

func (c *Client) GetLineageIntegration(
	projectID int64,
	lineageIntegrationID int64,
) (*LineageIntegration, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/integrations/lineage/%d/",
			c.HostURL,
			c.AccountID,
			projectID,
			lineageIntegrationID,
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

	lineageIntegrationResponse := LineageIntegrationResponse{}
	err = json.Unmarshal(body, &lineageIntegrationResponse)
	if err != nil {
		return nil, err
	}

	return &lineageIntegrationResponse.Data, nil
}

func (c *Client) CreateLineageIntegration(
	projectID int64,
	name string,
	host string,
	siteID string,
	tokenName string,
	token string,
) (*LineageIntegration, error) {
	newLineageIntegration := LineageIntegration{
		AccountID: int64(c.AccountID),
		ProjectID: projectID,
		Name:      name,
		Config: LineageIntegrationConfig{
			Host:      host,
			SiteID:    siteID,
			TokenName: tokenName,
			Token:     token,
		},
	}
	newLineageIntegrationData, err := json.Marshal(newLineageIntegration)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/integrations/lineage/",
			c.HostURL,
			c.AccountID,
			projectID,
		),
		strings.NewReader(string(newLineageIntegrationData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	lineageIntegrationResponse := LineageIntegrationResponse{}
	err = json.Unmarshal(body, &lineageIntegrationResponse)
	if err != nil {
		return nil, err
	}

	return &lineageIntegrationResponse.Data, nil
}

func (c *Client) UpdateLineageIntegration(
	projectID int64,
	lineageIntegrationID int64,
	lineageIntegration LineageIntegration,
) (*LineageIntegration, error) {
	lineageIntegrationData, err := json.Marshal(lineageIntegration)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/integrations/lineage/%d/",
			c.HostURL,
			c.AccountID,
			projectID,
			lineageIntegrationID,
		),
		strings.NewReader(string(lineageIntegrationData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	lineageIntegrationResponse := LineageIntegrationResponse{}
	err = json.Unmarshal(body, &lineageIntegrationResponse)
	if err != nil {
		return nil, err
	}

	return &lineageIntegrationResponse.Data, nil
}

func (c *Client) DeleteLineageIntegration(projectID int64, lineageIntegrationID int64) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/integrations/lineage/%d/",
			c.HostURL,
			c.AccountID,
			projectID,
			lineageIntegrationID,
		),
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithRetry(req)
	if err != nil {
		return err
	}

	return nil
}
