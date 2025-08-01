package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type FullEnvironmentVariable struct {
	Name                  string
	ProjectID             int
	EnvironmentNameValues map[string]EnvironmentVariableNameValue
}

type AbstractedEnvironmentVariable struct {
	Name              string
	ProjectID         int
	EnvironmentValues map[string]string
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

func (c *Client) GetEnvironmentVariable(
	projectID int,
	environmentVariableName string,
) (*FullEnvironmentVariable, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/environment-variables/environment/",
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

	environmentVariableResponse := GetEnvironmentVariableResponse{}
	err = json.Unmarshal(body, &environmentVariableResponse)
	if err != nil {
		return nil, err
	}

	environmentsVariables, _ := environmentVariableResponse.Data.Variables[environmentVariableName]
	if environmentsVariables == nil {
		return nil, fmt.Errorf(
			"resource-not-found: Environment variables %s not found in project ID %d",
			environmentVariableName,
			projectID,
		)
	}

	environmentVariable := FullEnvironmentVariable{
		Name:                  environmentVariableName,
		ProjectID:             projectID,
		EnvironmentNameValues: environmentsVariables,
	}

	return &environmentVariable, nil
}

func (c *Client) CreateEnvironmentVariable(
	projectID int,
	name string,
	environmentValues map[string]string,
) (*AbstractedEnvironmentVariable, error) {
	environmentValues["new_name"] = name
	newEnvironmentVariable := map[string]map[string]string{
		"env_var": environmentValues,
	}

	newEnvironmentVariableData, err := json.Marshal(newEnvironmentVariable)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/environment-variables/bulk/",
			c.HostURL,
			c.AccountID,
			projectID,
		),
		strings.NewReader(string(newEnvironmentVariableData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	createEnvironmentVariableResponse := CreateEnvironmentVariableResponse{}
	err = json.Unmarshal(body, &createEnvironmentVariableResponse)
	if err != nil {
		return nil, err
	}

	environmentVariable := AbstractedEnvironmentVariable{
		ProjectID:         projectID,
		Name:              name,
		EnvironmentValues: environmentValues,
	}
	return &environmentVariable, nil
}

func (c *Client) UpdateEnvironmentVariable(
	projectID int,
	environmentVariable AbstractedEnvironmentVariable,
) (*AbstractedEnvironmentVariable, error) {
	updateData := map[string]string{"name": environmentVariable.Name}
	for key, environmentVariableValue := range environmentVariable.EnvironmentValues {
		updateData[key] = environmentVariableValue
	}
	envVarUpdateData := map[string]map[string]string{"env_vars": updateData}

	environmentVariableData, err := json.Marshal(envVarUpdateData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/environment-variables/bulk/",
			c.HostURL,
			c.AccountID,
			projectID,
		),
		strings.NewReader(string(environmentVariableData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	createEnvironmentVariableResponse := CreateEnvironmentVariableResponse{}
	err = json.Unmarshal(body, &createEnvironmentVariableResponse)
	if err != nil {
		return nil, err
	}

	environmentVariable = AbstractedEnvironmentVariable{
		Name:              environmentVariable.Name,
		ProjectID:         projectID,
		EnvironmentValues: environmentVariable.EnvironmentValues,
	}

	return &environmentVariable, nil
}

func (c *Client) DeleteEnvironmentVariable(
	environmentVariableName string,
	projectID int,
) (string, error) {

	if environmentVariableName == "" || projectID <= 0 || c.AccountID <= 0 {
		return "", fmt.Errorf("invalid parameters: environmentVariableName = %s, projectID = %d, accountID = %d", environmentVariableName, projectID, c.AccountID)
	}

	environmentVariableData, _ := json.Marshal(map[string]string{"name": environmentVariableName})
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/environment-variables/bulk/",
			c.HostURL,
			c.AccountID,
			projectID,
		),
		strings.NewReader(string(environmentVariableData)),
	)
	if err != nil {
		return "", err
	}

	_, err = c.doRequestWithRetry(req)
	if err != nil {
		return "", err
	}

	return "", err
}
