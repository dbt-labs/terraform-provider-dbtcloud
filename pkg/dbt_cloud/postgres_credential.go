package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type PostgresCredential struct {
	ID             *int   `json:"id"`
	Account_Id     int    `json:"account_id"`
	Project_Id     int    `json:"project_id"`
	Type           string `json:"type"`
	State          int    `json:"state"`
	Threads        int    `json:"threads"`
	Username       string `json:"username"`
	Default_Schema string `json:"default_schema"`
	Target_Name    string `json:"target_name"`
	Password       string `json:"password,omitempty"`
}

type PostgresCredentialListResponse struct {
	Data   []PostgresCredential `json:"data"`
	Status ResponseStatus       `json:"status"`
}

type PostgresCredentialResponse struct {
	Data   PostgresCredential `json:"data"`
	Status ResponseStatus     `json:"status"`
}

// GetPostgresCredential retrieves a specific Postgres credential by its ID
func (c *Client) GetPostgresCredential(projectId int, credentialId int) (*PostgresCredential, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/credentials/", c.HostURL, c.AccountID, projectId), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	PostgresCredentialListResponse := PostgresCredentialListResponse{}
	err = json.Unmarshal(body, &PostgresCredentialListResponse)
	if err != nil {
		return nil, err
	}

	for i, credential := range PostgresCredentialListResponse.Data {
		if *credential.ID == credentialId {
			return &PostgresCredentialListResponse.Data[i], nil
		}
	}

	return nil, fmt.Errorf("did not find credential ID %d in project ID %d", credentialId, projectId)
}

// CreatePostgresCredential creates a new Postgres credential
func (c *Client) CreatePostgresCredential(projectId int, isActive bool, type_ string, defaultSchema string, targetName string, username string, password string, numThreads int) (*PostgresCredential, error) {
	newPostgresCredential := PostgresCredential{
		Account_Id:     c.AccountID,
		Project_Id:     projectId,
		Type:           type_,
		State:          STATE_ACTIVE, // TODO: make variable
		Threads:        numThreads,
		Username:       username,
		Default_Schema: defaultSchema,
		Target_Name:    targetName,
		Password:       password,
	}
	newPostgresCredentialData, err := json.Marshal(newPostgresCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/credentials/", c.HostURL, c.AccountID, projectId), strings.NewReader(string(newPostgresCredentialData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	PostgresCredentialResponse := PostgresCredentialResponse{}
	err = json.Unmarshal(body, &PostgresCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &PostgresCredentialResponse.Data, nil
}

// UpdatePostgresCredential updates an existing Postgres credential
func (c *Client) UpdatePostgresCredential(projectId int, credentialId int, postgresCredential PostgresCredential) (*PostgresCredential, error) {
	postgresCredentialData, err := json.Marshal(postgresCredential)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/credentials/%d/", c.HostURL, c.AccountID, projectId, credentialId), strings.NewReader(string(postgresCredentialData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	postgresCredentialResponse := PostgresCredentialResponse{}
	err = json.Unmarshal(body, &postgresCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &postgresCredentialResponse.Data, nil
}

// DeletePostgresCredential deletes a Postgres credential by its ID
func (c *Client) DeletePostgresCredential(credentialId, projectId string) (string, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v3/accounts/%d/projects/%s/credentials/%s/", c.HostURL, c.AccountID, projectId, credentialId), nil)
	if err != nil {
		return "", err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return "", err
	}

	return "", err
}
