package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SnowflakeCredentialResponse struct {
	Data   SnowflakeCredential `json:"data"`
	Status ResponseStatus      `json:"status"`
}

type SnowflakeCredential struct {
	ID                   *int   `json:"id"`
	Account_Id           int    `json:"account_id"`
	Project_Id           int    `json:"project_id"`
	Type                 string `json:"type"`
	State                int    `json:"state"`
	Threads              int    `json:"threads"`
	User                 string `json:"user"`
	Password             string `json:"password,omitempty"`
	Auth_Type            string `json:"auth_type"`
	Database             string `json:"database"`
	Role                 string `json:"role"`
	Warehouse            string `json:"warehouse"`
	Schema               string `json:"schema"`
	PrivateKey           string `json:"private_key,omitempty"`
	PrivateKeyPassphrase string `json:"private_key_passphrase,omitempty"`
}

func (c *Client) GetSnowflakeCredential(
	projectId int,
	credentialId int,
) (*SnowflakeCredential, error) {
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

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	snowflakeCredentialResponse := SnowflakeCredentialResponse{}
	err = json.Unmarshal(body, &snowflakeCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &snowflakeCredentialResponse.Data, nil
}

func (c *Client) CreateSnowflakeCredential(
	projectId int,
	type_ string,
	isActive bool,
	database string,
	role string,
	warehouse string,
	schema string,
	user string,
	password string,
	privateKey string,
	privateKeyPassphrase string,
	authType string,
	numThreads int,
) (*SnowflakeCredential, error) {
	newSnowflakeCredential := SnowflakeCredential{
		Account_Id: c.AccountID,
		Project_Id: projectId,
		Type:       type_,
		State:      STATE_ACTIVE,
		Database:   database,
		Role:       role,
		Warehouse:  warehouse,
		Schema:     schema,
		User:       user,
		Auth_Type:  authType,
		Threads:    numThreads,
	}
	if authType == "password" {
		newSnowflakeCredential.Password = password
	}
	if authType == "keypair" {
		newSnowflakeCredential.PrivateKey = privateKey
		newSnowflakeCredential.PrivateKeyPassphrase = privateKeyPassphrase
	}
	newSnowflakeCredentialData, err := json.Marshal(newSnowflakeCredential)
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
		strings.NewReader(string(newSnowflakeCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	snowflakeCredentialResponse := SnowflakeCredentialResponse{}
	err = json.Unmarshal(body, &snowflakeCredentialResponse)
	if err != nil {
		return nil, err
	}
	if authType == "password" {
		snowflakeCredentialResponse.Data.Password = password
	}
	if authType == "keypair" {
		snowflakeCredentialResponse.Data.PrivateKey = privateKey
		snowflakeCredentialResponse.Data.PrivateKeyPassphrase = privateKeyPassphrase
	}

	return &snowflakeCredentialResponse.Data, nil
}

func (c *Client) UpdateSnowflakeCredential(
	projectId int,
	credentialId int,
	snowflakeCredential SnowflakeCredential,
) (*SnowflakeCredential, error) {
	snowflakeCredentialData, err := json.Marshal(snowflakeCredential)
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
		strings.NewReader(string(snowflakeCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	snowflakeCredentialResponse := SnowflakeCredentialResponse{}
	err = json.Unmarshal(body, &snowflakeCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &snowflakeCredentialResponse.Data, nil
}
