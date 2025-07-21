package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type BigQueryCredentialResponse struct {
	Data   BigQueryCredential `json:"data"`
	Status ResponseStatus     `json:"status"`
}

type BigQueryCredential struct {
	ID         *int   `json:"id"`
	Account_Id int    `json:"account_id"`
	Project_Id int    `json:"project_id"`
	Type       string `json:"type"`
	State      int    `json:"state"`
	Threads    int    `json:"threads"`
	Dataset    string `json:"schema"`
}

func (c *Client) GetBigQueryCredential(
	projectId int,
	credentialId int,
) (*BigQueryCredential, error) {
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

	BigQueryCredentialResponse := BigQueryCredentialResponse{}
	err = json.Unmarshal(body, &BigQueryCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &BigQueryCredentialResponse.Data, nil
}

func (c *Client) CreateBigQueryCredential(
	projectId int,
	type_ string,
	isActive bool,
	dataset string,
	numThreads int,
) (*BigQueryCredential, error) {
	newBigQueryCredential := BigQueryCredential{
		Account_Id: c.AccountID,
		Project_Id: projectId,
		Type:       type_,
		State:      STATE_ACTIVE,
		Dataset:    dataset,
		Threads:    numThreads,
	}
	newBigQueryCredentialData, err := json.Marshal(newBigQueryCredential)
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
		strings.NewReader(string(newBigQueryCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	BigQueryCredentialResponse := BigQueryCredentialResponse{}
	err = json.Unmarshal(body, &BigQueryCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &BigQueryCredentialResponse.Data, nil
}

func (c *Client) UpdateBigQueryCredential(
	projectId int,
	credentialId int,
	BigQueryCredential BigQueryCredential,
) (*BigQueryCredential, error) {
	BigQueryCredentialData, err := json.Marshal(BigQueryCredential)
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
		strings.NewReader(string(BigQueryCredentialData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	BigQueryCredentialResponse := BigQueryCredentialResponse{}
	err = json.Unmarshal(body, &BigQueryCredentialResponse)
	if err != nil {
		return nil, err
	}

	return &BigQueryCredentialResponse.Data, nil
}
