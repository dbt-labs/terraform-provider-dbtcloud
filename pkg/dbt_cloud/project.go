package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Project struct {
	ID    *int   `json:"id"`
	Name  string `json:"name"`
	State int    `json:"state"`
}

type ProjectResponse struct {
	Data Project `json:"data"`
}

func (c *Client) GetProject(projectID string) (*Project, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/%s/", c.AccountURL, projectID), nil)
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
