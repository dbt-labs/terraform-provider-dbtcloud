package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SnowflakeCredentialListResponse struct {
	Data   []SnowflakeCredential `json:"data"`
	Status ResponseStatus        `json:"status"`
}

type SnowflakeCredentialResponse struct {
	Data   SnowflakeCredential `json:"data"`
	Status ResponseStatus      `json:"status"`
}

type SnowflakeCredential struct {
	ID         *int   `json:"id"`
	Account_Id int    `json:"account_id"`
	Project_Id int    `json:"project_id"`
	Type       string `json:"type"`
	State      int    `json:"state"`
	Threads    int    `json:"threads"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Auth_Type  string `json:"auth_type"`
	Schema     string `json:"schema"`
}

func (c *Client) GetSnowflakeCredential(projectId int, credentialId int) (*SnowflakeCredential, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/credentials/", HostURL, c.AccountID, projectId), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	snowflakeCredentialListResponse := SnowflakeCredentialListResponse{}
	err = json.Unmarshal(body, &snowflakeCredentialListResponse)
	if err != nil {
		return nil, err
	}

	for i, credential := range snowflakeCredentialListResponse.Data {
		if *credential.ID == credentialId {
			return &snowflakeCredentialListResponse.Data[i], nil
		}
	}

	return nil, fmt.Errorf("did not find credential ID %d in project ID %d", credentialId, projectId)
}

func (c *Client) CreateSnowflakeCredential(projectId int, type_ string, isActive bool, schema string, user string, password string, authType string, numThreads int) (*SnowflakeCredential, error) {
	newSnowflakeCredential := SnowflakeCredential{
		Account_Id: c.AccountID,
		Project_Id: projectId,
		Type:       type_,
		State:      1, // TODO: make variable
		Schema:     schema,
		User:       user,
		Password:   password,
		Auth_Type:  authType,
		Threads:    numThreads,
	}
	newSnowflakeCredentialData, err := json.Marshal(newSnowflakeCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/credentials/", HostURL, c.AccountID, projectId), strings.NewReader(string(newSnowflakeCredentialData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
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

func (c *Client) UpdateSnowflakeCredential(projectId int, credentialId int, snowflakeCredential SnowflakeCredential) (*SnowflakeCredential, error) {
	snowflakeCredentialData, err := json.Marshal(snowflakeCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/credentials/%d", HostURL, c.AccountID, projectId, credentialId), strings.NewReader(string(snowflakeCredentialData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
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
