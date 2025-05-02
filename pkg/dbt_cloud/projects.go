package dbt_cloud

import (
	"encoding/json"
	"fmt"
)

type ProjectConnectionRepository struct {
	Name                   string                                `json:"name,omitempty"`
	AccountID              int64                                 `json:"account_id,omitempty"`
	Description            string                                `json:"description,omitempty"`
	ConnectionID           int64                                 `json:"connection_id,omitempty"`
	RepositoryID           int64                                 `json:"repository_id,omitempty"`
	SemanticLayerConfigID  *int64                                `json:"semantic_layer_config_id,omitempty"`
	SkippedSetup           bool                                  `json:"skipped_setup,omitempty"`
	State                  int64                                 `json:"state,omitempty"`
	DbtProjectSubdirectory string                                `json:"dbt_project_subdirectory,omitempty"`
	DocsJobID              *int64                                `json:"docs_job_id,omitempty"`
	DbtProjectType         int64                                 `json:"type"`
	FreshnessJobID         *int64                                `json:"freshness_job_id,omitempty"`
	ID                     int64                                 `json:"id,omitempty"`
	CreatedAt              string                                `json:"created_at,omitempty"`
	UpdatedAt              string                                `json:"updated_at,omitempty"`
	Connection             *globalConnectionPayload[EmptyConfig] `json:"connection,omitempty"`
	Environments           any                                   `json:"environments,omitempty"`
	Repository             *Repository                           `json:"repository,omitempty"`
	GroupPermissions       any                                   `json:"group_permissions,omitempty"`
	DocsJob                any                                   `json:"docs_job,omitempty"`
	FreshnessJob           any                                   `json:"freshness_job,omitempty"`
}

func (c *Client) GetAllProjects(nameContains string) ([]ProjectConnectionRepository, error) {
	var url string

	if nameContains == "" {
		url = fmt.Sprintf(
			`%s/v3/accounts/%d/projects/?limit=100&order_by=name&include_related=["repository","connection"]`,
			c.HostURL,
			c.AccountID,
		)
	} else {
		url = fmt.Sprintf(
			`%s/v3/accounts/%d/projects/?name__icontains=%s&limit=100&order_by=name&include_related=["repository","connection"]`,
			c.HostURL,
			c.AccountID,
			nameContains,
		)
	}

	allProjectsRaw := c.GetData(url)

	allProjects := []ProjectConnectionRepository{}
	for _, job := range allProjectsRaw {

		data, _ := json.Marshal(job)
		currentProject := ProjectConnectionRepository{}
		err := json.Unmarshal(data, &currentProject)
		if err != nil {
			return nil, err
		}
		allProjects = append(allProjects, currentProject)
	}
	return allProjects, nil
}
