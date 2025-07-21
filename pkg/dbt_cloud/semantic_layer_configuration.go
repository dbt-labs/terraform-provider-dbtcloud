package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type SemanticLayerConfiguration struct {
	ID            int64  `json:"id,omitempty"`
	AccountID     int64  `json:"account_id"`
	ProjectID     int64  `json:"project_id"`
	EnvironmentID int64  `json:"environment_id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	State         int64  `json:"state"`
}

type SemanticLayerConfigurationResponse struct {
	Data   SemanticLayerConfiguration `json:"data"`
	Status ResponseStatus             `json:"status"`
}

func (c *Client) GetSemanticLayerConfiguration(projectId int64, semanticLayerConfigId int64) (*SemanticLayerConfiguration, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/semantic-layer-configurations/%s/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			strconv.Itoa(int(projectId)),
			strconv.Itoa(int(semanticLayerConfigId)),
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

	configResponse := SemanticLayerConfigurationResponse{}
	err = json.Unmarshal(body, &configResponse)
	if err != nil {
		return nil, err
	}

	return &configResponse.Data, nil
}

func (c *Client) CreateSemanticLayerConfiguration(
	projectId int64,
	environmentId int64,
) (*SemanticLayerConfiguration, error) {

	newConfig := SemanticLayerConfiguration{
		AccountID:     int64(c.AccountID),
		ProjectID:     projectId,
		EnvironmentID: environmentId,
		State:         1,
	}

	newConfigData, err := json.Marshal(newConfig)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/semantic-layer-configurations/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			strconv.Itoa(int(projectId)),
		),
		strings.NewReader(string(newConfigData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	configResponse := SemanticLayerConfigurationResponse{}
	err = json.Unmarshal(body, &configResponse)
	if err != nil {
		return nil, err
	}

	return &configResponse.Data, nil
}

func (c *Client) UpdateSemanticLayerConfiguration(
	projectId int64,
	semanticLayerConfigId int64,
	semanticLayerConfig SemanticLayerConfiguration) (*SemanticLayerConfiguration, error) {

	configData, err := json.Marshal(semanticLayerConfig)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/semantic-layer-configurations/%s/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			strconv.Itoa(int(projectId)),
			strconv.Itoa(int(semanticLayerConfigId)),
		),
		strings.NewReader(string(configData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	configResponse := SemanticLayerConfigurationResponse{}
	err = json.Unmarshal(body, &configResponse)
	if err != nil {
		return nil, err
	}

	return &configResponse.Data, nil
}

func (c *Client) DeleteSemanticLayerConfiguration(
	projectId int64,
	semanticLayerConfigurationID int64,
) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/semantic-layer-configurations/%s/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			strconv.Itoa(int(projectId)),
			strconv.Itoa(int(semanticLayerConfigurationID)),
		),
		nil,
	)
	if err != nil {
		return err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return err
	}

	configResponse := SemanticLayerConfigurationResponse{}
	err = json.Unmarshal(body, &configResponse)
	if err != nil {
		return err
	}

	return nil
}
