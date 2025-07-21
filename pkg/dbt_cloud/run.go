package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Run struct {
	ID                  int64  `json:"id,omitempty"`
	AccountID           int64  `json:"account_id"`
	JobID               int    `json:"job_id"`
	GitSHA              string `json:"git_sha,omitempty"`
	GitBranch           string `json:"git_branch,omitempty"`
	GitHubPullRequestID string `json:"github_pull_request_id,omitempty"`
	SchemaOverride      string `json:"schema_override,omitempty"`
	Cause               string `json:"cause,omitempty"`
}

type RunResponse struct {
	Data   Run            `json:"data"`
	Status ResponseStatus `json:"status"`
}

type RunsResponse struct {
	Data   []Run          `json:"data"`
	Status ResponseStatus `json:"status"`
	Extra  ResponseExtra  `json:"extra"`
}

type RunFilter struct {
	Limit           int    `json:"limit"`
	EnvironmentID   int    `json:"environment_id"`
	ProjectID       int    `json:"project_id"`
	TriggerID       int    `json:"trigger_id"`
	JobDefinitionID int    `json:"job_definition_id"`
	PullRequestID   int    `json:"pull_request_id"`
	Status          int    `json:"status"`
	StatusIn        string `json:"status_in"`
}

func (c *Client) GetRun(runID int64) (*Run, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v2/accounts/%s/runs/%s/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			strconv.Itoa(int(runID)),
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

	runResponse := RunResponse{}
	err = json.Unmarshal(body, &runResponse)
	if err != nil {
		return nil, err
	}

	return &runResponse.Data, nil
}

func (c *Client) GetRuns(filter *RunFilter) (*[]Run, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v2/accounts/%s/runs/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Add query parameters if a filter is provided
	query := req.URL.Query()
	if filter != nil {
		if filter.Status != 0 {
			query.Add("status", strconv.Itoa(filter.Status))
		}
		if filter.Limit > 0 {
			query.Add("limit", strconv.Itoa(filter.Limit))
		}
		if filter.EnvironmentID > 0 {
			query.Add("environment_id", strconv.Itoa(filter.EnvironmentID))
		}
		if filter.ProjectID > 0 {
			query.Add("project_id", strconv.Itoa(filter.ProjectID))
		}
		if filter.TriggerID > 0 {
			query.Add("trigger_id", strconv.Itoa(filter.TriggerID))
		}
		if filter.JobDefinitionID > 0 {
			query.Add("job_definition_id", strconv.Itoa(filter.JobDefinitionID))
		}
		if filter.PullRequestID > 0 {
			query.Add("pull_request_id", strconv.Itoa(filter.PullRequestID))
		}
		if filter.StatusIn != "" {
			query.Add("status_in", filter.StatusIn)
		}
	}
	req.URL.RawQuery = query.Encode()

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response into the appropriate struct
	var response RunsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *Client) TriggerRun(
	jobID int,
	gitSHA string,
	gitBranch string,
	githubPullRequestID string,
	schemaOverride string) (*Run, error) {

	newRun := Run{
		AccountID:           int64(c.AccountID),
		JobID:               jobID,
		GitSHA:              gitSHA,
		GitBranch:           gitBranch,
		GitHubPullRequestID: githubPullRequestID,
		SchemaOverride:      schemaOverride,
		Cause:               "API",
	}

	newRunData, err := json.Marshal(newRun)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v2/accounts/%s/jobs/%s/run/",
			c.HostURL,
			strconv.Itoa(int(jobID)),
			strconv.Itoa(c.AccountID),
		),
		strings.NewReader(string(newRunData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	runResponse := RunResponse{}
	err = json.Unmarshal(body, &runResponse)
	if err != nil {
		return nil, err
	}

	return &runResponse.Data, nil
}

func (c *Client) CancelRun(runID int64) (*Run, error) {

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v2/accounts/%s/runs/%s/cancel",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			strconv.Itoa(int(runID)),
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

	runResponse := RunResponse{}
	err = json.Unmarshal(body, &runResponse)
	if err != nil {
		return nil, err
	}

	return &runResponse.Data, nil
}

func (c *Client) RetryRun(runID int64) (*Run, error) {

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v2/accounts/%s/runs/%s/retry",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			strconv.Itoa(int(runID)),
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

	runResponse := RunResponse{}
	err = json.Unmarshal(body, &runResponse)
	if err != nil {
		return nil, err
	}

	return &runResponse.Data, nil
}
