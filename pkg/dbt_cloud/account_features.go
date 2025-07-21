package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type AccountFeaturesResponse struct {
	Data   AccountFeatures `json:"data"`
	Status ResponseStatus  `json:"status"`
	Extra  ResponseExtra   `json:"extra"`
}

type AccountFeatures struct {
	AdvancedCI              bool `json:"advanced-ci"`
	PartialParsing          bool `json:"partial-parsing"`
	RepoCaching             bool `json:"repo-caching"`
	AIFeatures              bool `json:"ai_features"`
	WarehouseCostVisibility bool `json:"warehouse_cost_visibility"`
}

type AccountFeatureUpdateRequest struct {
	Feature string `json:"feature"`
	Value   bool   `json:"value"`
}

func (c *Client) GetAccountFeatures() (*AccountFeatures, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/private/accounts/%d/features/", c.HostURL, c.AccountID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	featuresResponse := AccountFeaturesResponse{}
	err = json.Unmarshal(body, &featuresResponse)
	if err != nil {
		return nil, err
	}

	return &featuresResponse.Data, nil
}

func (c *Client) UpdateAccountFeature(feature string, value bool) error {
	updateRequest := AccountFeatureUpdateRequest{
		Feature: feature,
		Value:   value,
	}

	updateData, err := json.Marshal(updateRequest)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/private/accounts/%d/features/", c.HostURL, c.AccountID),
		strings.NewReader(string(updateData)),
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithRetry(req)
	return err
}
