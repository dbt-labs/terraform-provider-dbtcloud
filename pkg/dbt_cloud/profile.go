package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ProfileResponse struct {
	Data   Profile        `json:"data"`
	Status ResponseStatus `json:"status"`
}

type ProfileListResponse struct {
	Data   []Profile      `json:"data"`
	Status ResponseStatus `json:"status"`
}

type Profile struct {
	ID                   *int   `json:"id,omitempty"`
	AccountID            int    `json:"account_id"`
	ProjectID            int    `json:"project_id"`
	Key                  string `json:"key"`
	ConnectionID         int    `json:"connection_id"`
	CredentialsID        int    `json:"credentials_id"`
	ExtendedAttributesID *int   `json:"extended_attributes_id"`
}

func (c *Client) GetProfile(projectID int, profileID int) (*Profile, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/profiles/%d/",
			c.HostURL,
			c.AccountID,
			projectID,
			profileID,
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

	profileResponse := ProfileResponse{}
	err = json.Unmarshal(body, &profileResponse)
	if err != nil {
		return nil, err
	}

	return &profileResponse.Data, nil
}

func (c *Client) GetAllProfiles(projectID int) ([]Profile, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/profiles/",
			c.HostURL,
			c.AccountID,
			projectID,
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

	profileListResponse := ProfileListResponse{}
	err = json.Unmarshal(body, &profileListResponse)
	if err != nil {
		return nil, err
	}

	return profileListResponse.Data, nil
}

func (c *Client) CreateProfile(
	projectID int,
	key string,
	connectionID int,
	credentialsID int,
	extendedAttributesID *int,
) (*Profile, error) {
	newProfile := Profile{
		AccountID:            c.AccountID,
		ProjectID:            projectID,
		Key:                  key,
		ConnectionID:         connectionID,
		CredentialsID:        credentialsID,
		ExtendedAttributesID: extendedAttributesID,
	}

	newProfileData, err := json.Marshal(newProfile)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/profiles/",
			c.HostURL,
			c.AccountID,
			projectID,
		),
		strings.NewReader(string(newProfileData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	profileResponse := ProfileResponse{}
	err = json.Unmarshal(body, &profileResponse)
	if err != nil {
		return nil, err
	}

	return &profileResponse.Data, nil
}

func (c *Client) UpdateProfile(
	projectID int,
	profileID int,
	profile Profile,
) (*Profile, error) {
	profileData, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PATCH",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/profiles/%d/",
			c.HostURL,
			c.AccountID,
			projectID,
			profileID,
		),
		strings.NewReader(string(profileData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	profileResponse := ProfileResponse{}
	err = json.Unmarshal(body, &profileResponse)
	if err != nil {
		return nil, err
	}

	return &profileResponse.Data, nil
}

func (c *Client) DeleteProfile(projectID int, profileID int) (string, error) {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/profiles/%d/",
			c.HostURL,
			c.AccountID,
			projectID,
			profileID,
		),
		nil,
	)
	if err != nil {
		return "", err
	}

	_, err = c.doRequestWithRetry(req)
	if err != nil {
		return "", err
	}

	return "", nil
}
