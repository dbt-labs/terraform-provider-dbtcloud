package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type EnvironmentVariableJobOverride struct {
	ID              *int   `json:"id"`
	AccountID       int    `json:"account_id"`
	JobDefinitionID int    `json:"job_definition_id"`
	Name            string `json:"name"`
	ProjectID       int    `json:"project_id"`
	RawValue        string `json:"raw_value"`
	Type            string `json:"type"`
}

type EnvironmentVariableJobOverrideResponse struct {
	Data   EnvironmentVariableJobOverride `json:"data"`
	Status ResponseStatus                 `json:"status"`
}

type EnvironmentVariableJobOverrideAllResponse struct {
	Data   any            `json:"data"`
	Status ResponseStatus `json:"status"`
}

func (c *Client) GetEnvironmentVariableJobOverride(
	projectID int,
	jobDefinitionID int,
	environmentVariableOverrideID int,
) (*EnvironmentVariableJobOverride, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/environment-variables/job/?job_definition_id=%d",
			c.HostURL,
			c.AccountID,
			projectID,
			jobDefinitionID,
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

	environmentVariableJobOverrideAllResponse := EnvironmentVariableJobOverrideAllResponse{}
	err = json.Unmarshal(body, &environmentVariableJobOverrideAllResponse)
	if err != nil {
		return nil, err
	}

	dataMap := environmentVariableJobOverrideAllResponse.Data.(map[string]interface{})

	for envVarName, value := range dataMap {
		innerMap, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("could not unpack the data")
		}

		// the default is to be a float64 when we unmarshall a generic interface{}
		jobMap, ok := innerMap["job"].(map[string]interface{})

		if ok {

			if overrideID, ok := jobMap["id"].(float64); ok &&
				int(overrideID) == environmentVariableOverrideID {

				environmentVariableJobOverride := EnvironmentVariableJobOverride{
					AccountID:       int(c.AccountID),
					Name:            envVarName,
					ProjectID:       projectID,
					RawValue:        jobMap["value"].(string),
					Type:            "job",
					JobDefinitionID: jobDefinitionID,
					ID:              &environmentVariableOverrideID,
				}

				return &environmentVariableJobOverride, nil
			}
		}

	}

	return nil, fmt.Errorf(
		"resource-not-found: Did not find the override %d",
		environmentVariableOverrideID,
	)
}

func (c *Client) CreateEnvironmentVariableJobOverride(
	projectID int,
	name string,
	rawValue string,
	jobDefinitionID int,
) (*EnvironmentVariableJobOverride, error) {

	envOverride := EnvironmentVariableJobOverride{
		AccountID:       int(c.AccountID),
		Name:            name,
		ProjectID:       projectID,
		RawValue:        rawValue,
		Type:            "job",
		JobDefinitionID: jobDefinitionID,
		ID:              nil,
	}

	envOverrideData, err := json.Marshal(envOverride)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/environment-variables/",
			c.HostURL,
			c.AccountID,
			projectID,
		),
		strings.NewReader(string(envOverrideData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	environmentVariableJobOverrideResponse := EnvironmentVariableJobOverrideResponse{}
	err = json.Unmarshal(body, &environmentVariableJobOverrideResponse)
	if err != nil {
		return nil, err
	}

	return &environmentVariableJobOverrideResponse.Data, nil
}

func (c *Client) UpdateEnvironmentVariableJobOverride(
	projectID int,
	environmentVariableJobOverrideID int,
	environmentVariableJobOverride EnvironmentVariableJobOverride,
) (*EnvironmentVariableJobOverride, error) {

	envOverrideData, err := json.Marshal(environmentVariableJobOverride)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/environment-variables/%d/",
			c.HostURL,
			c.AccountID,
			projectID,
			environmentVariableJobOverrideID,
		),
		strings.NewReader(string(envOverrideData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	environmentVariableJobOverrideResponse := EnvironmentVariableJobOverrideResponse{}
	err = json.Unmarshal(body, &environmentVariableJobOverrideResponse)
	if err != nil {
		return nil, err
	}

	return &environmentVariableJobOverrideResponse.Data, nil
}

func (c *Client) DeleteEnvironmentVariableJobOverride(
	projectID int,
	environmentVariableJobOverrideID int,
) (string, error) {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%d/projects/%d/environment-variables/%d/",
			c.HostURL,
			c.AccountID,
			projectID,
			environmentVariableJobOverrideID,
		),
		nil,
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
