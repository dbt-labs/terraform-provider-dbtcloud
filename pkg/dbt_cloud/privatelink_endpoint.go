package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PrivatelinkEndpoint struct {
	Account_Id             int    `json:"account_id"`
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
		return nil, fmt.Errorf("The endpoint name or url needs to be provided")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%d/private-link-endpoints/", c.HostURL, c.AccountID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	PrivatelinkEndpointListResponse := PrivatelinkEndpointListResponse{}
	err = json.Unmarshal(body, &PrivatelinkEndpointListResponse)
	if err != nil {
		return nil, err
	}

	for i, endpoint := range PrivatelinkEndpointListResponse.Data {
		if (endpointName == "" || endpoint.Name == endpointName) &&
			(privatelinkEndpointURL == "" || endpoint.PrivatelinkEndpointURL == privatelinkEndpointURL) {
			return &PrivatelinkEndpointListResponse.Data[i], nil
		}
	}

	return nil, fmt.Errorf("Did not find PrivateLink endpoint with name = '%s' and/or endpoint = '%s'", endpointName, privatelinkEndpointURL)
}

func (c *Client) GetAllPrivatelinkEndpoints() ([]PrivatelinkEndpoint, error) {
	url := fmt.Sprintf("%s/v3/accounts/%d/privatelink-endpoints/", c.HostURL, c.AccountID)

	allPrivatelinkEndpointsRaw := c.GetData(url)

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
