package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Repository struct {
	ID                                    *int       `json:"id,omitempty"`
	AccountID                             int64      `json:"account_id"`
	ProjectID                             int        `json:"project_id"`
	RemoteUrl                             string     `json:"remote_url"`
	State                                 int        `json:"state"`
	AzureActiveDirectoryProjectID         *string    `json:"azure_active_directory_project_id,omitempty"`
	AzureActiveDirectoryRepositoryID      *string    `json:"azure_active_directory_repository_id,omitempty"`
	AzureBypassWebhookRegistrationFailure *bool      `json:"azure_bypass_webhook_registration_failure,omitempty"`
	GitCloneStrategy                      string     `json:"git_clone_strategy"`
	RepositoryCredentialsID               *int       `json:"repository_credentials_id,omitempty"`
	GitlabProjectID                       *int       `json:"gitlab_project_id,omitempty"`
	GithubInstallationID                  *int       `json:"github_installation_id,omitempty"`
	PrivateLinkEndpointID                 *string    `json:"private_link_endpoint_id,omitempty"`
	DeployKey                             *DeployKey `json:"deploy_key,omitempty"`
	DeployKeyID                           *int       `json:"deploy_key_id,omitempty"`
	PullRequestURLTemplate                string     `json:"pull_request_url_template,omitempty"`
	RemoteBackend                         *string    `json:"remote_backend,omitempty"`
	FullName                              *string    `json:"full_name,omitempty"`
}

type DeployKey struct {
	ID        int    `json:"id"`
	AccountID int64  `json:"account_id"`
	State     int    `json:"state"`
	PublicKey string `json:"public_key"`
}

type RepositoryListResponse struct {
	Data   []Repository   `json:"data"`
	Status ResponseStatus `json:"status"`
}

type RepositoryResponse struct {
	Data   Repository     `json:"data"`
	Status ResponseStatus `json:"status"`
}

func (c *Client) GetRepository(
	repositoryID, projectID string,
) (*Repository, error) {

	repositoryUrl := fmt.Sprintf(
		"%s/v3/accounts/%s/projects/%s/repositories/%s/",
		c.HostURL,
		strconv.FormatInt(c.AccountID, 10),
		projectID,
		repositoryID,
	)

	req, err := http.NewRequest("GET", repositoryUrl, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	repositoryResponse := RepositoryResponse{}
	err = json.Unmarshal(body, &repositoryResponse)
	if err != nil {
		return nil, err
	}

	return &repositoryResponse.Data, nil
}

func (c *Client) CreateRepository(
	projectID int,
	remoteUrl string,
	isActive bool,
	gitCloneStrategy string,
	gitlabProjectID int,
	githubInstallationID int,
	privateLinkEndpointID string,
	azureActiveDirectoryProjectID string,
	azureActiveDirectoryRepositoryID string,
	azureBypassWebhookRegistrationFailure bool,
	pullRequestURLTemplate string,
) (*Repository, error) {
	state := STATE_ACTIVE
	if !isActive {
		state = STATE_DELETED
	}

	newRepository := Repository{
		AccountID:              c.AccountID,
		ProjectID:              projectID,
		RemoteUrl:              remoteUrl,
		State:                  state,
		GitCloneStrategy:       gitCloneStrategy,
		PullRequestURLTemplate: pullRequestURLTemplate,
	}
	if gitlabProjectID != 0 {
		newRepository.GitlabProjectID = &gitlabProjectID
	}
	if githubInstallationID != 0 {
		newRepository.GithubInstallationID = &githubInstallationID
	}
	if privateLinkEndpointID != "" {
		newRepository.PrivateLinkEndpointID = &privateLinkEndpointID
	}
	if azureActiveDirectoryProjectID != "" {
		newRepository.AzureActiveDirectoryProjectID = &azureActiveDirectoryProjectID
		newRepository.AzureActiveDirectoryRepositoryID = &azureActiveDirectoryRepositoryID
		newRepository.AzureBypassWebhookRegistrationFailure = &azureBypassWebhookRegistrationFailure
	}
	newRepositoryData, err := json.Marshal(newRepository)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/repositories/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			strconv.Itoa(projectID),
		),
		strings.NewReader(string(newRepositoryData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	repositoryResponse := RepositoryResponse{}
	err = json.Unmarshal(body, &repositoryResponse)
	if err != nil {
		return nil, err
	}

	if pullRequestURLTemplate != "" {
		// this is odd but we can't provide the pullRequestURLTemplate in the initial create
		// we need to update the repository with the pullRequestURLTemplate

		// we need to update the repository with the deploy key id that was created
		if repositoryResponse.Data.DeployKeyID != nil {
			newRepository.DeployKeyID = repositoryResponse.Data.DeployKeyID
		}
		// and we also need to provide the credentials id if it was created
		if repositoryResponse.Data.RepositoryCredentialsID != nil {
			newRepository.RepositoryCredentialsID = repositoryResponse.Data.RepositoryCredentialsID
		}

		if repositoryResponse.Data.RemoteBackend != nil {
			newRepository.RemoteBackend = repositoryResponse.Data.RemoteBackend
		}

		if repositoryResponse.Data.FullName != nil {
			newRepository.FullName = repositoryResponse.Data.FullName
		}

		updatedRepo, err := c.UpdateRepository(
			strconv.Itoa(*repositoryResponse.Data.ID),
			strconv.Itoa(projectID),
			newRepository,
		)
		if err != nil {
			return nil, err
		}
		return updatedRepo, nil
	}

	return &repositoryResponse.Data, nil
}

func (c *Client) UpdateRepository(
	repositoryID, projectID string,
	repository Repository,
) (*Repository, error) {

	// we need to remove the GitLab project ID for updates
	repository.GitlabProjectID = nil

	repositoryData, err := json.Marshal(repository)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/repositories/%s/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			projectID,
			repositoryID,
		),
		strings.NewReader(string(repositoryData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	repositoryResponse := RepositoryResponse{}
	err = json.Unmarshal(body, &repositoryResponse)
	if err != nil {
		return nil, err
	}

	return &repositoryResponse.Data, nil
}

func (c *Client) DeleteRepository(repositoryID, projectID string) (string, error) {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/repositories/%s/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			projectID,
			repositoryID,
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
