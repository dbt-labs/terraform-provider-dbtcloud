package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type EnvironmentVariable struct {
	Name                  string
	ProjectID             int
	EnvironmentNameValues map[string]string
}

type EnvironmentVariableNameValue struct {
	ID    int    `json:"id,omitempty"`
	Value string `json:"value"`
}

type EnvironmentVariablesGet struct {
	Environment []string                                           `json:"environments"`
	Variables   map[string]map[string]EnvironmentVariableNameValue `json:"variables"`
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

func (c *Client) GetEnvironmentVariable(projectID int, environmentVariableName string) (*EnvironmentVariable, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environment-variables/environment", c.HostURL, c.AccountID, projectID), nil)
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

	environmentsVariables := environmentVariableResponse.Data.Variables[environmentVariableName]
	if environmentsVariables == nil {
	    return nil, fmt.Errorf("did not find environment variables %s in project ID %d", environmentVariableName, projectID)
	}
	environmentValues := make(map[string]string)
	for environmentName, environmentVariableNameValue := range environmentsVariables {
		environmentValues[environmentName] = environmentVariableNameValue.Value
	}

	environmentVariable := EnvironmentVariable{
		Name:                  environmentVariableName,
		ProjectID:             projectID,
		EnvironmentNameValues: environmentValues,
	}

	return &environmentVariable, nil
}

func (c *Client) CreateEnvironmentVariable(projectID int, name string, environmentValues map[string]string) (*EnvironmentVariable, error) {
	environmentValues["new_name"] = name
	newEnvironmentVariable := map[string]map[string]string{
		"env_var": environmentValues,
	}

	newEnvironmentVariableData, err := json.Marshal(newEnvironmentVariable)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environment-variables/bulk/", c.HostURL, c.AccountID, projectID), strings.NewReader(string(newEnvironmentVariableData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	createEnvironmentVariableResponse := CreateEnvironmentVariableResponse{}
	err = json.Unmarshal(body, &createEnvironmentVariableResponse)
	if err != nil {
		return nil, err
	}

	environmentVariable := EnvironmentVariable{
		ProjectID:             projectID,
		Name:                  name,
		EnvironmentNameValues: environmentValues,
	}
	return &environmentVariable, nil
}

func (c *Client) UpdateEnvironmentVariable(projectID int, environmentVariable EnvironmentVariable) (*EnvironmentVariable, error) {
	updateData := map[string]string{"name": environmentVariable.Name}
	for environmentName, environmentVariableValue := range environmentVariable.EnvironmentNameValues {
		updateData[environmentName] = environmentVariableValue
	}
	envVarUpdateData := map[string]map[string]string{"env_vars": updateData}

	environmentVariableData, err := json.Marshal(envVarUpdateData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/projects/%d/environment-variables/bulk/", c.HostURL, c.AccountID, projectID), strings.NewReader(string(environmentVariableData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	createEnvironmentVariableResponse := CreateEnvironmentVariableResponse{}
	err = json.Unmarshal(body, &createEnvironmentVariableResponse)
	if err != nil {
		return nil, err
	}

	newEnvVariables := map[string]string{}
	for environmentName, environmentVariableValue := range environmentVariable.EnvironmentNameValues {
		newEnvVariables[environmentName] = environmentVariableValue
	}

	environmentVariable = EnvironmentVariable{
		Name:                  environmentVariable.Name,
		ProjectID:             projectID,
		EnvironmentNameValues: newEnvVariables,
	}

	return &environmentVariable, nil
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
