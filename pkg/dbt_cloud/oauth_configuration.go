package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type OAuthConfiguration struct {
	ID                      *int64                   `json:"id,omitempty"`
	AccountId               int64                    `json:"account_id"`
	Type                    string                   `json:"type"`
	Name                    string                   `json:"name"`
	ClientId                string                   `json:"client_id"`
	ClientSecret            string                   `json:"client_secret"`
	AuthorizeUrl            string                   `json:"authorize_url"`
	TokenUrl                string                   `json:"token_url"`
	RedirectUri             string                   `json:"redirect_uri"`
	OAuthConfigurationExtra *OAuthConfigurationExtra `json:"extra_data,omitempty"`
}

type OAuthConfigurationExtra struct {
	ApplicationIdUri *string `json:"application_id_uri,omitempty"`
}

type OAuthConfigurationListResponse struct {
	Data   []OAuthConfiguration `json:"data"`
	Status ResponseStatus       `json:"status"`
	Extra  ResponseExtra        `json:"extra"`
}

type OAuthConfigurationResponse struct {
	Data   OAuthConfiguration `json:"data"`
	Status ResponseStatus     `json:"status"`
}

func (c *Client) GetOAuthConfiguration(oAuthConfigurationID int64) (*OAuthConfiguration, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/oauth-configurations/%d/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			oAuthConfigurationID,
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

	oAuthConfigurationResponse := OAuthConfigurationResponse{}
	err = json.Unmarshal(body, &oAuthConfigurationResponse)
	if err != nil {
		return nil, err
	}

	return &oAuthConfigurationResponse.Data, nil
}

func (c *Client) CreateOAuthConfiguration(
	oAuthType string,
	name string,
	clientId string,
	clientSecret string,
	authorizeUrl string,
	tokenUrl string,
	redirectUri string,
	applicationURI string,
) (*OAuthConfiguration, error) {
	newOAuthConfiguration := OAuthConfiguration{
		AccountId:    int64(c.AccountID),
		Type:         oAuthType,
		Name:         name,
		ClientId:     clientId,
		ClientSecret: clientSecret,
		AuthorizeUrl: authorizeUrl,
		TokenUrl:     tokenUrl,
		RedirectUri:  redirectUri,
	}

	if applicationURI != "" {
		newOAuthConfigurationExtra := OAuthConfigurationExtra{
			ApplicationIdUri: &applicationURI,
		}
		newOAuthConfiguration.OAuthConfigurationExtra = &newOAuthConfigurationExtra
	}

	newOAuthConfigurationData, err := json.Marshal(newOAuthConfiguration)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/oauth-configurations/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
		),
		strings.NewReader(string(newOAuthConfigurationData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	oAuthConfigurationResponse := OAuthConfigurationResponse{}
	err = json.Unmarshal(body, &oAuthConfigurationResponse)
	if err != nil {
		return nil, err
	}

	return &oAuthConfigurationResponse.Data, nil
}

func (c *Client) UpdateOAuthConfiguration(
	oAuthConfigurationID int64,
	oAuthConfiguration OAuthConfiguration,
) (*OAuthConfiguration, error) {
	oAuthConfigurationData, err := json.Marshal(oAuthConfiguration)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/oauth-configurations/%d/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			oAuthConfigurationID,
		),
		strings.NewReader(string(oAuthConfigurationData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	oAuthConfigurationResponse := OAuthConfigurationResponse{}
	err = json.Unmarshal(body, &oAuthConfigurationResponse)
	if err != nil {
		return nil, err
	}

	return &oAuthConfigurationResponse.Data, nil
}

func (c *Client) DeleteOAuthConfiguration(
	oAuthConfigurationID int64,
) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%s/oauth-configurations/%d/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			oAuthConfigurationID,
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

	oAuthConfigurationResponse := OAuthConfigurationResponse{}
	err = json.Unmarshal(body, &oAuthConfigurationResponse)
	if err != nil {
		return err
	}

	return nil
}
