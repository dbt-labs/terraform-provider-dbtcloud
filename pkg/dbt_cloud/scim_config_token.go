package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type SCIMConfigToken struct {
	ID          *int64  `json:"id,omitempty"`
	Name        string  `json:"name"`
	CreatedAt   string  `json:"created_at,omitempty"`
	LastUsed    *string `json:"last_used,omitempty"`
	TokenString *string `json:"token_string,omitempty"`
}

type SCIMConfigTokenResponse struct {
	Data   SCIMConfigToken `json:"data"`
	Status ResponseStatus  `json:"status"`
}

type SCIMConfigTokenListResponse struct {
	Data   []SCIMConfigToken `json:"data"`
	Status ResponseStatus    `json:"status"`
}

func (c *Client) GetSCIMConfigToken(tokenID int64) (*SCIMConfigToken, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/scim-config/tokens/%d/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			tokenID,
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

	resp := SCIMConfigTokenResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) CreateSCIMConfigToken(name string) (*SCIMConfigToken, error) {
	payload, err := json.Marshal(SCIMConfigToken{Name: name})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/scim-config/tokens/",
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

	resp := SCIMConfigTokenResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) DeleteSCIMConfigToken(tokenID int64) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%s/scim-config/tokens/%d/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			tokenID,
		),
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithRetry(req)
	return err
}
