package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type EnvironmentResponse struct {
	Data   Environment    `json:"data"`
	Status ResponseStatus `json:"status"`
}

type Environment struct {
	ID                           *int                 `json:"id,omitempty"`
	State                        int                  `json:"state,omitempty"`
	Account_Id                   int                  `json:"account_id"`
	Project_Id                   int                  `json:"project_id"`
	Credential_Id                *int                 `json:"credentials_id,omitempty"`
	Name                         string               `json:"name"`
	Dbt_Version                  string               `json:"dbt_version"`
	Type                         string               `json:"type"`
	Use_Custom_Branch            bool                 `json:"use_custom_branch"`
	Custom_Branch                *string              `json:"custom_branch"`
	Environment_Id               *int                 `json:"-"`
	Support_Docs                 bool                 `json:"supports_docs"`
	Created_At                   *string              `json:"created_at"`
	Updated_At                   *string              `json:"updated_at"`
	Project                      Project              `json:"project"`
	Jobs                         *string              `json:"jobs"`
	Credentials                  *SnowflakeCredential `json:"credentials"`
	Custom_Environment_Variables *string              `json:"custom_environment_variables"`
	DeploymentType               *string              `json:"deployment_type,omitempty"`
}

func (c *Client) GetEnvironment(projectId int, environmentId int) (*Environment, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environments/%d/", c.HostURL, c.AccountID, projectId, environmentId), nil)
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

	environmentResponse.Data.Environment_Id = &environmentId
	return &environmentResponse.Data, nil
}

func (c *Client) CreateEnvironment(isActive bool, projectId int, name string, dbtVersion string, type_ string, useCustomBranch bool, customBranch string, credentialId int, deploymentType string) (*Environment, error) {
	state := STATE_ACTIVE
	if !isActive {
		state = STATE_DELETED
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
	if deploymentType != "" {
		newEnvironment.DeploymentType = &deploymentType
	}
	newEnvironmentData, err := json.Marshal(newEnvironment)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environments/", c.HostURL, c.AccountID, projectId), strings.NewReader(string(newEnvironmentData)))
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

	// we don't send the environment details in the update request, just the credential_id
	environment.Credentials = nil

	environmentData, err := json.Marshal(environment)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environments/%d/", c.HostURL, c.AccountID, projectId, environmentId), strings.NewReader(string(environmentData)))
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

func (c *Client) DeleteEnvironment(projectId, environmentId int) (string, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environments/%d/", c.HostURL, c.AccountID, projectId, environmentId), nil)
	if err != nil {
		return "", err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return "", err
	}

	return "", err
}
