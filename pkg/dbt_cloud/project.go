package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Project struct {
	ID                     *int    `json:"id,omitempty"`
	Name                   string  `json:"name"`
	DbtProjectSubdirectory *string `json:"dbt_project_subdirectory,omitempty"`
	ConnectionID           *int    `json:"connection_id,integer,omitempty"`
	RepositoryID           *int    `json:"repository_id,integer,omitempty"`
	State                  int     `json:"state"`
	AccountID              int     `json:"account_id"`
}

type ProjectListResponse struct {
	Data   []Project      `json:"data"`
	Status ResponseStatus `json:"status"`
}

type ProjectResponse struct {
	Data   Project        `json:"data"`
	Status ResponseStatus `json:"status"`
}

func (c *Client) GetProject(projectID string) (*Project, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%s/projects/", c.HostURL, strconv.Itoa(c.AccountID)), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	projectListResponse := ProjectListResponse{}
	err = json.Unmarshal(body, &projectListResponse)
	if err != nil {
		return nil, err
	}
	
	for i, project := range projectListResponse.Data {
		if strconv.Itoa(*project.ID) == projectID {
 			return &projectListResponse.Data[i], nil
 		}
 	}

	return nil, fmt.Errorf("Did not find project ID %d in account ID %d", projectID, strconv.Itoa(c.AccountID))
}

func (c *Client) CreateProject(name string, dbtProjectSubdirectory string, connectionID int, repositoryID int) (*Project, error) {
	newProject := Project{
		Name:                   name,
		State:                  1,
		AccountID:              c.AccountID,
	}
	if dbtProjectSubdirectory != "" {
		newProject.DbtProjectSubdirectory = &dbtProjectSubdirectory
	}
	if connectionID != 0 {
		newProject.ConnectionID = &connectionID
	}
	if repositoryID != 0 {
		newProject.RepositoryID = &repositoryID
	}
	
	newProjectData, err := json.Marshal(newProject)
	if err != nil {
		return nil, err
	}
	log.Println(string(newProjectData))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/projects/", c.HostURL, strconv.Itoa(c.AccountID)), strings.NewReader(string(newProjectData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	projectResponse := ProjectResponse{}
	err = json.Unmarshal(body, &projectResponse)
	if err != nil {
		return nil, err
	}

	return &projectResponse.Data, nil
}

func (c *Client) UpdateProject(projectID string, project Project) (*Project, error) {
	projectData, err := json.Marshal(project)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/projects/%s/", c.HostURL, strconv.Itoa(c.AccountID), projectID), strings.NewReader(string(projectData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	projectResponse := ProjectResponse{}
	err = json.Unmarshal(body, &projectResponse)
	if err != nil {
		return nil, err
	}

	return &projectResponse.Data, nil
}
