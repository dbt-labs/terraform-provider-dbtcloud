package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type RedshiftCredentialResponse struct {
	Data   RedshiftCredential `json:"data"`
	Status ResponseStatus     `json:"status"`
}

type RedshiftCredential struct {
	ID            *int   `json:"id"`
	Account_Id    int    `json:"account_id"`
	Project_Id    int    `json:"project_id"`
	Type          string `json:"type"`
	State         int    `json:"state"`
	Threads       int    `json:"threads"`
	DefaultSchema string `json:"default_schema"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

func (c *Client) GetRedshiftCredential(
	projectId int,
	credentialId int,
) (*RedshiftCredential, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/credentials/%d/",
			c.HostURL,
			c.AccountID,
			projectId,
			credentialId,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	RedshiftCredentialResponse := RedshiftCredentialResponse{}
	err = json.Unmarshal(body, &RedshiftCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &RedshiftCredentialResponse.Data, nil
}

func (c *Client) CreateRedshiftCredential(
	projectId int,
	type_ string,
	isActive bool,
	schema string,
	numThreads int,
	username string,
	password string,
) (*RedshiftCredential, error) {
	newRedshiftCredential := RedshiftCredential{
		Account_Id:    c.AccountID,
		Project_Id:    projectId,
		Type:          type_,
		State:         STATE_ACTIVE,
		Threads:       numThreads,
		DefaultSchema: schema,
		Username:      username,
		Password:      password,
	}

	newRedshiftCredentialData, err := json.Marshal(newRedshiftCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/credentials/",
			c.HostURL,
			c.AccountID,
			projectId,
		),
		strings.NewReader(string(newRedshiftCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	RedshiftCredentialResponse := RedshiftCredentialResponse{}
	err = json.Unmarshal(body, &RedshiftCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &RedshiftCredentialResponse.Data, nil
}

func (c *Client) UpdateRedshiftCredential(
	projectId int,
	credentialId int,
	RedshiftCredential RedshiftCredential,
) (*RedshiftCredential, error) {
	RedshiftCredentialData, err := json.Marshal(RedshiftCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/credentials/%d/",
			c.HostURL,
			c.AccountID,
			projectId,
			credentialId,
		),
		strings.NewReader(string(RedshiftCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	RedshiftCredentialResponse := RedshiftCredentialResponse{}
	err = json.Unmarshal(body, &RedshiftCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &RedshiftCredentialResponse.Data, nil
}
