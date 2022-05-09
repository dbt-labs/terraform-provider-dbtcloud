package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Repository struct {
	ID                      *int   `json:"id,omitempty"`
	AccountID               int    `json:"account_id"`
	ProjectID               int    `json:"project_id"`
	RemoteUrl               string `json:"remote_url"`
	State                   int    `json:"state"`
	GitCloneStrategy        string `json:"git_clone_strategy"`
	RepositoryCredentialsID *int   `json:"repository_credentials_id"`
	GitlabProjectID         *int   `json:"gitlab_project_id"`
}

type RepositoryListResponse struct {
	Data   []Repository   `json:"data"`
	Status ResponseStatus `json:"status"`
}

type RepositoryResponse struct {
	Data   Repository     `json:"data"`
	Status ResponseStatus `json:"status"`
}

func (c *Client) GetRepository(repositoryID, projectID string) (*Repository, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/repositories/%s/", c.HostURL, strconv.Itoa(c.AccountID), projectID, repositoryID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
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

func (c *Client) CreateRepository(projectID int, remoteUrl string, isActive bool, gitCloneStrategy string, repositoryCredentialsID int, gitlabProjectID int) (*Repository, error) {
	state := STATE_ACTIVE
	if !isActive {
		state = STATE_DELETED
	}

	newRepository := Repository{
		AccountID:        c.AccountID,
		ProjectID:        projectID,
		RemoteUrl:        remoteUrl,
		State:            state,
		GitCloneStrategy: gitCloneStrategy,
	}
	if repositoryCredentialsID != 0 {
		newRepository.RepositoryCredentialsID = &repositoryCredentialsID
	}
	if gitlabProjectID != 0 {
		newRepository.GitlabProjectID = &gitlabProjectID
	}

	newRepositoryData, err := json.Marshal(newRepository)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/repositories/", c.HostURL, strconv.Itoa(c.AccountID), strconv.Itoa(projectID)), strings.NewReader(string(newRepositoryData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
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

func (c *Client) UpdateRepository(repositoryID, projectID string, repository Repository) (*Repository, error) {
	repositoryData, err := json.Marshal(repository)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/repositories/%s/", c.HostURL, strconv.Itoa(c.AccountID), projectID, repositoryID), strings.NewReader(string(repositoryData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
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
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/repositories/%s/", c.HostURL, strconv.Itoa(c.AccountID), projectID, repositoryID), nil)
	if err != nil {
		return "", err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return "", err
	}

	return "", err
}
