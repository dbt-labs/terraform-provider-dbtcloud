package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type EnvironmentVariable struct {
	ID    int
	Value string
	Name string
	EnvironmentName string
}

type EnvironmentVariablesGet struct {
	Environment []string                                  `json:"environments"`
	Variables   map[string]map[string]EnvironmentVariable `json:"variables"`
}

type GetEnvironmentVariableResponse struct {
	Data   EnvironmentVariablesGet `json:"data"`
	Status ResponseStatus          `json:"status"`
}

type CreateEnvironmentVariableResponseMessage struct {
	Message        string `json:"message"`
	NewVariableIDs []int  `json:"new_var_ids"`
}
type CreateEnvironmentVariableResponse struct {
	Data   CreateEnvironmentVariableResponseMessage `json:"data"`
	Status ResponseStatus                           `json:"status"`
}

func (c *Client) GetEnvironmentVariable(projectID int, variableKey string) (*EnvironmentVariable, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environment-variables/environment", c.HostURL, c.AccountID, projectId), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	environmentVariableResponse := GetEnvironmentVariableResponse{}
	err = json.Unmarshal(body, &environmentVariableResponse)
	if err != nil {
		return nil, err
	}

	return &environmentResponse.Data.Variables[variableKey], nil
}

func (c *Client) CreateEnvironmentVariable(projectID int, Name string, EnvironmentName string, Value string) (*EnvironmentVariable, error) {
	EnvironmentValues["new_name"] = Name
	newEnvironmentVariable := map[string]map[string]string{
		"env_var": EnvironmentValues,
	}

	newEnvironmentVariableData, err := json.Marshal(newEnvironmentVariable)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environment-variables/bulk/", c.HostURL, c.AccountID, projectId), strings.NewReader(string(newEnvironmentVariableData)))
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

	return &environmentResponse.Data, nil
}

func (c *Client) UpdateEnvironment(projectId int, environmentId int, environment Environment) (*Environment, error) {
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

func (c *Client) DeleteEnvironmentVariable(environmentVariableName string, projectID int) (string, error) {
    environmentVariableData, err := json.Marshal(map[string]string{"name": environmentVariableName})
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environment-variables/bulk/", c.HostURL, c.AccountID, projectID), strings.NewReader(string(environmentVariableData)))
	if err != nil {
		return "", err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return "", err
	}

	return "", err
}
