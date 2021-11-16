package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type EnvironmentListResponse struct {
	Data   []Environment  `json:"data"`
	Status ResponseStatus `json:"status"`
}

type EnvironmentResponse struct {
	Data   Environment    `json:"data"`
	Status ResponseStatus `json:"status"`
}

type Environment struct {
	ID                *int    `json:"id,omitempty"`
	State             int     `json:"state,omitempty"`
	Account_Id        int     `json:"account_id"`
	Project_Id        int     `json:"project_id"`
	Credential_Id     *int    `json:"credentials_id,omitempty"`
	Name              string  `json:"name"`
	Dbt_Version       string  `json:"dbt_version"`
	Type              string  `json:"type"`
	Use_Custom_Branch bool    `json:"use_custom_branch"`
	Custom_Branch     *string `json:"custom_branch,omitempty"`
	Environment_Id    *int    `json:"environment_id,omitempty"`
}

func (c *Client) GetEnvironment(projectId int, environmentId int) (*Environment, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environments/", HostURL, c.AccountID, projectId), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	environmentListResponse := EnvironmentListResponse{}
	err = json.Unmarshal(body, &environmentListResponse)
	if err != nil {
		return nil, err
	}

	for i, environment := range environmentListResponse.Data {
		if *environment.ID == environmentId {
			environmentListResponse.Data[i].Environment_Id = &environmentId
			return &environmentListResponse.Data[i], nil
		}
	}

	return nil, fmt.Errorf("did not find environment ID %d in project ID %d", environmentId, projectId)
}

func (c *Client) CreateEnvironment(isActive bool, projectId int, name string, dbtVersion string, type_ string, useCustomBranch bool, customBranch string, credentialId int) (*Environment, error) {
	state := 1
	if !isActive {
		state = 2
	}

	newEnvironment := Environment{
		State:             state,
		Account_Id:        c.AccountID,
		Project_Id:        projectId,
		Name:              name,
		Dbt_Version:       dbtVersion,
		Type:              type_,
		Use_Custom_Branch: useCustomBranch,
	}
	if credentialId != 0 {
		newEnvironment.Credential_Id = &credentialId
	}
	if customBranch != "" {
		newEnvironment.Custom_Branch = &customBranch
	}
	newEnvironmentData, err := json.Marshal(newEnvironment)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environments/", HostURL, c.AccountID, projectId), strings.NewReader(string(newEnvironmentData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	environmentResponse := EnvironmentResponse{}
	err = json.Unmarshal(body, &environmentResponse)
	if err != nil {
		return nil, err
	}

	environmentResponse.Data.Environment_Id = environmentResponse.Data.ID
	return &environmentResponse.Data, nil
}

func (c *Client) UpdateEnvironment(projectId int, environmentId int, environment Environment) (*Environment, error) {
	environmentData, err := json.Marshal(environment)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environments/%d", HostURL, c.AccountID, projectId, environmentId), strings.NewReader(string(environmentData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	environmentResponse := EnvironmentResponse{}
	err = json.Unmarshal(body, &environmentResponse)
	if err != nil {
		return nil, err
	}

	environmentResponse.Data.Environment_Id = environmentResponse.Data.ID
	return &environmentResponse.Data, nil
}
