package dbt_cloud

import (
	"encoding/json"
	"fmt"
)

type PrivatelinkEndpoint struct {
	Account_Id             int64  `json:"account_id"`
	Name                   string `json:"name"`
	Type                   string `json:"type"`
	PrivatelinkEndpointURL string `json:"private_link_endpoint"`
	CIDRRange              string `json:"cidr_range"`
	State                  int    `json:"state"`
	ID                     string `json:"id"`
}

type PrivatelinkEndpointListResponse struct {
	Data   []PrivatelinkEndpoint `json:"data"`
	Status ResponseStatus        `json:"status"`
}

type PrivatelinkEndpointResponse struct {
	Data   PrivatelinkEndpoint `json:"data"`
	Status ResponseStatus      `json:"status"`
}

func (c *Client) GetPrivatelinkEndpoint(endpointName string, privatelinkEndpointURL string) (*PrivatelinkEndpoint, error) {

	if endpointName == "" && privatelinkEndpointURL == "" {
		return nil, fmt.Errorf("the endpoint name or url needs to be provided")
	}

	url := c.BuildAccountV3URL(ResourcePrivatelinkEndpoints)

	allPrivatelinkEndpointsRaw, err := c.GetRawData(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw data for PrivateLink endpoints: %w", err)
	}

	for _, privatelinkEndpointRaw := range allPrivatelinkEndpointsRaw {
		data, _ := json.Marshal(privatelinkEndpointRaw)
		currentPrivatelinkEndpoint := PrivatelinkEndpoint{}
		err := json.Unmarshal(data, &currentPrivatelinkEndpoint)
		if err != nil {
			return nil, err
		}

		if (endpointName == "" || currentPrivatelinkEndpoint.Name == endpointName) &&
			(privatelinkEndpointURL == "" || currentPrivatelinkEndpoint.PrivatelinkEndpointURL == privatelinkEndpointURL) {
			return &currentPrivatelinkEndpoint, nil
		}
	}

	return nil, fmt.Errorf("did not find PrivateLink endpoint with name = '%s' and/or endpoint = '%s'", endpointName, privatelinkEndpointURL)
}

func (c *Client) GetAllPrivatelinkEndpoints() ([]PrivatelinkEndpoint, error) {
	url := c.BuildAccountV3URL(ResourcePrivatelinkEndpoints)

	allPrivatelinkEndpointsRaw, err := c.GetRawData(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw data for all PrivateLink endpoints: %w", err)
	}

	allPrivatelinkEndpoints := []PrivatelinkEndpoint{}
	for _, privatelinkEndpoint := range allPrivatelinkEndpointsRaw {
		data, _ := json.Marshal(privatelinkEndpoint)
		currentPrivatelinkEndpoint := PrivatelinkEndpoint{}
		err := json.Unmarshal(data, &currentPrivatelinkEndpoint)
		if err != nil {
			return nil, err
		}
		allPrivatelinkEndpoints = append(allPrivatelinkEndpoints, currentPrivatelinkEndpoint)
	}

	return allPrivatelinkEndpoints, nil
}
