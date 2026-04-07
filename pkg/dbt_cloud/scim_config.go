package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type SCIMConfig struct {
	Enabled                   bool `json:"enabled"`
	ManualUpdatesAllowed      bool `json:"manual_updates_allowed"`
	SCIMControlledLicenseType bool `json:"scim_controlled_license_type"`
}

type SCIMConfigResponse struct {
	Data   SCIMConfig     `json:"data"`
	Status ResponseStatus `json:"status"`
}

func (c *Client) GetSCIMConfig() (*SCIMConfig, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/scim-config/",
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

	resp := SCIMConfigResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) UpdateSCIMConfig(cfg SCIMConfig) (*SCIMConfig, error) {
	payload, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/scim-config/",
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

	resp := SCIMConfigResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}
