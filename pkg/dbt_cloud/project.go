package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Project struct {
	ID                     *int    `json:"id,omitempty"`
	Name                   string  `json:"name"`
	Description            string  `json:"description"`
	DbtProjectSubdirectory *string `json:"dbt_project_subdirectory,omitempty"`
	DbtProjectType         int64   `json:"type"`
	ConnectionID           *int    `json:"connection_id,omitempty"`
	RepositoryID           *int    `json:"repository_id,omitempty"`
	State                  int     `json:"state"`
	AccountID              int     `json:"account_id"`
	FreshnessJobId         *int    `json:"freshness_job_id"`
	DocsJobId              *int    `json:"docs_job_id,"`
	SemanticLayerConfigID  *int64  `json:"semantic_layer_config_id,omitempty"`
}

type ProjectListResponse struct {
	Data   []Project      `json:"data"`
	Status ResponseStatus `json:"status"`
	Extra  ResponseExtra  `json:"extra"`
}

type ProjectResponse struct {
	Data   Project        `json:"data"`
	Status ResponseStatus `json:"status"`
}

const InvalidFileCharacters = `#%&{}<>*?$!'":@`

func (c *Client) GetProjectByName(projectName string) (*Project, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/?include_related=[freshness_job_id,docs_job_id]",
			c.HostURL,
			strconv.Itoa(c.AccountID),
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

	projectListResponse := ProjectListResponse{}
	err = json.Unmarshal(body, &projectListResponse)
	if err != nil {
		return nil, err
	}

	listAllProjects := projectListResponse.Data

	// if there are more than the limit, we need to paginate
	if projectListResponse.Extra.Pagination.TotalCount > projectListResponse.Extra.Filters.Limit {
		numProjects := projectListResponse.Extra.Pagination.Count
		for numProjects < projectListResponse.Extra.Pagination.TotalCount {

			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf(
					"%s/v3/accounts/%s/projects/?include_related=[freshness_job_id,docs_job_id]&offset=%d",
					c.HostURL,
					strconv.Itoa(c.AccountID),
					numProjects,
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

			projectListResponse := ProjectListResponse{}
			err = json.Unmarshal(body, &projectListResponse)
			if err != nil {
				return nil, err
			}

			numProjectsLastCall := projectListResponse.Extra.Pagination.Count
			if numProjectsLastCall > 0 {
				listAllProjects = append(listAllProjects, projectListResponse.Data...)
				numProjects += projectListResponse.Extra.Pagination.Count
			} else {
				// this means that most likely one item was deleted since the first call
				// so the number of items is less than the initial total, we can break the loop
				break
			}

		}
	}

	// we now loop though the projects to find the ones with the name we are looking for
	matchingProjects := []Project{}
	for _, project := range listAllProjects {
		if strings.EqualFold(project.Name, projectName) {
			matchingProjects = append(matchingProjects, project)
		}
	}

	if len(matchingProjects) == 0 {
		return nil, fmt.Errorf("Did not find any project with the name: %s", projectName)
	} else if len(matchingProjects) > 1 {
		return nil, fmt.Errorf("Found more than one project with the name: %s", projectName)
	}

	return &matchingProjects[0], nil
}

func (c *Client) GetProject(projectID string) (*Project, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/?include_related=[freshness_job_id,docs_job_id]",
			c.HostURL,
			strconv.Itoa(c.AccountID),
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

	projectResponse := ProjectResponse{}
	err = json.Unmarshal(body, &projectResponse)
	if err != nil {
		return nil, err
	}

	// Validate the response has required fields
	if err := ValidateProjectResponse(projectResponse.Data.ID, &projectResponse.Data); err != nil {
		return nil, err
	}

	return &projectResponse.Data, nil
}

func (c *Client) CreateProject(
	name string,
	description string,
	dbtProjectSubdirectory string,
	dbtProjectType int64,
) (*Project, error) {
	newProject := Project{
		Name:           name,
		Description:    description,
		State:          STATE_ACTIVE,
		AccountID:      c.AccountID,
		DbtProjectType: dbtProjectType,
	}

	dbtProjectSubdirectory = strings.TrimSpace(dbtProjectSubdirectory)
	if dbtProjectSubdirectory != "" {
		if err := IsValidSubdirectory(dbtProjectSubdirectory); err != nil {
			return nil, err
		}
		newProject.DbtProjectSubdirectory = &dbtProjectSubdirectory
	}

	newProjectData, err := json.Marshal(newProject)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/v3/accounts/%s/projects/", c.HostURL, strconv.Itoa(c.AccountID)),
		strings.NewReader(string(newProjectData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	projectResponse := ProjectResponse{}
	err = json.Unmarshal(body, &projectResponse)
	if err != nil {
		return nil, err
	}

	// Validate the response has required fields
	if err := ValidateProjectResponse(projectResponse.Data.ID, &projectResponse.Data); err != nil {
		return nil, err
	}

	return &projectResponse.Data, nil
}

func (c *Client) UpdateProject(projectID string, project Project) (*Project, error) {
	if project.DbtProjectSubdirectory != nil {
		*project.DbtProjectSubdirectory = strings.TrimSpace(*project.DbtProjectSubdirectory)
		if err := IsValidSubdirectory(*project.DbtProjectSubdirectory); err != nil {
			return nil, err
		}
	}

	projectData, err := json.Marshal(project)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			projectID,
		),
		strings.NewReader(string(projectData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	projectResponse := ProjectResponse{}
	err = json.Unmarshal(body, &projectResponse)
	if err != nil {
		return nil, err
	}

	// Validate the response has required fields
	if err := ValidateProjectResponse(projectResponse.Data.ID, &projectResponse.Data); err != nil {
		return nil, err
	}

	return &projectResponse.Data, nil
}

func IsValidSubdirectory(dbtProjectSubdirectory string) error {
	if strings.HasPrefix(dbtProjectSubdirectory, "/") {
		return fmt.Errorf(`project subdirectory path should not start with a slash: "%s"`, dbtProjectSubdirectory)
	}

	if strings.HasSuffix(dbtProjectSubdirectory, "/") {
		return fmt.Errorf(`project subdirectory path should not end with a slash: "%s"`, dbtProjectSubdirectory)
	}

	if strings.Contains(dbtProjectSubdirectory, "./") || strings.Contains(dbtProjectSubdirectory, "~/") {
		return fmt.Errorf(`project subdirectory path should not contain relative paths: "%s"`, dbtProjectSubdirectory)
	}

	if strings.ContainsAny(dbtProjectSubdirectory, InvalidFileCharacters) {
		return fmt.Errorf(`project subdirectory path should not contain file characters ("%s"): "%s"`, InvalidFileCharacters, dbtProjectSubdirectory)
	}

	return nil
}
