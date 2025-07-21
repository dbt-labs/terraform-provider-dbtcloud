package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AzureDevOpsRepository struct {
	DetailsURL    string `json:"url"`
	RemoteURL     string `json:"remoteUrl"`
	WebURL        string `json:"webUrl"`
	Name          string `json:"name"`
	ID            string `json:"id"`
	DefaultBranch string `json:"defaultBranch"`
}

type AzureDevOpsRepositoriesData struct {
	Count int                     `json:"count"`
	Value []AzureDevOpsRepository `json:"value"`
}

type AzureDevOpsRepositoriesResponse struct {
	Data   AzureDevOpsRepositoriesData `json:"data"`
	Status ResponseStatus              `json:"status"`
}

func (c *Client) GetAzureDevOpsRepositories(
	azureDevOpsProjectID string,
) ([]AzureDevOpsRepository, error) {

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/integrations/azure-ad/projects/%s/repositories/?account_id=%d",
			c.HostURL,
			azureDevOpsProjectID,
			c.AccountID,
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

	AzureDevOpsRepositoryListResponse := AzureDevOpsRepositoriesResponse{}
	err = json.Unmarshal(body, &AzureDevOpsRepositoryListResponse)
	if err != nil {
		return nil, err
	}

	return AzureDevOpsRepositoryListResponse.Data.Value, nil

}

func (c *Client) GetAzureDevOpsRepository(
	repositoryName string,
	azureDevOpsProjectID string,
) (*AzureDevOpsRepository, error) {

	listAzureDevOpsRepositories, err := c.GetAzureDevOpsRepositories(azureDevOpsProjectID)
	if err != nil {
		return nil, err
	}

	for _, adoRepository := range listAzureDevOpsRepositories {
		if adoRepository.Name == repositoryName {
			return &adoRepository, nil
		}
	}

	return nil, fmt.Errorf(
		"Did not find any Azure Dev Ops project with the name = '%s' in the ADO project with the ID = '%s'",
		repositoryName,
		azureDevOpsProjectID,
	)
}
