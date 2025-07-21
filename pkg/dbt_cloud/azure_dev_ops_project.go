package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AzureDevOpsProject struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	ID   string `json:"id"`
}

type AzureDevOpsProjectsData struct {
	Count int                  `json:"count"`
	Value []AzureDevOpsProject `json:"value"`
}

type AzureDevOpsProjectsResponse struct {
	Data   AzureDevOpsProjectsData `json:"data"`
	Status ResponseStatus          `json:"status"`
}

func (c *Client) GetAzureDevOpsProjects() ([]AzureDevOpsProject, error) {

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/v3/integrations/azure-ad/projects/?account_id=%d", c.HostURL, c.AccountID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	AzureDevOpsProjectListResponse := AzureDevOpsProjectsResponse{}
	err = json.Unmarshal(body, &AzureDevOpsProjectListResponse)
	if err != nil {
		return nil, err
	}

	return AzureDevOpsProjectListResponse.Data.Value, nil

}

func (c *Client) GetAzureDevOpsProject(
	projectName string,
) (*AzureDevOpsProject, error) {

	listAzureDevOpsProjects, err := c.GetAzureDevOpsProjects()
	if err != nil {
		return nil, err
	}

	for _, adoProject := range listAzureDevOpsProjects {
		if adoProject.Name == projectName {
			return &adoProject, nil
		}
	}

	return nil, fmt.Errorf(
		"Did not find any Azure Dev Ops project with the name = '%s'",
		projectName,
	)
}
