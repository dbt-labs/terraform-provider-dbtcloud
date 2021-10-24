package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type JobTrigger struct {
	Github_Webhook     bool `json:"github_webhook"`
	Schedule           bool `json:"schedule"`
	Custom_Branch_Only bool `json:"custom_branch_only"`
}

type JobSettings struct {
	Threads     int    `json:"threads"`
	Target_Name string `json:"target_name"`
}

type scheduleDate struct {
	Type string `json:"type"`
}

type scheduleTime struct {
	Type     string `json:"type"`
	Interval int    `json:"interval"`
}

type JobSchedule struct {
	Cron string       `json:"cron"`
	Date scheduleDate `json:"date"`
	Time scheduleTime `json:"time"`
}

type JobResponse struct {
	Data   Job            `json:"data"`
	Status ResponseStatus `json:"status"`
}

type Job struct {
	ID                   *int        `json:"id"`
	Account_Id           int         `json:"account_id"`
	Project_Id           int         `json:"project_id"`
	Environment_Id       int         `json:"environment_id"`
	Name                 string      `json:"name"`
	Execute_Steps        []string    `json:"execute_steps"`
	Dbt_Version          *string     `json:"dbt_version"`
	Triggers             JobTrigger  `json:"triggers"`
	Settings             JobSettings `json:"settings"`
	State                int         `json:"state"`
	Generate_Docs        bool        `json:"generate_docs"`
	Schedule             JobSchedule `json:"schedule"`
	Run_Generate_Sources bool        `json:"run_generate_sources"`
}

func (c *Client) GetJob(jobID string) (*Job, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/accounts/%s/v2/jobs/%s/", c.HostURL, c.AccountID, jobID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	jobResponse := JobResponse{}
	err = json.Unmarshal(body, &jobResponse)
	if err != nil {
		return nil, err
	}

	return &jobResponse.Data, nil
}

func (c *Client) CreateJob(projectId int, environmentId int, name string, executeSteps []string, dbtVersion string, isActive bool, triggers map[string]interface{}, numThreads int, targetName string, generateDocs bool, runGenerateSources bool) (*Job, error) {
	state := 1
	if !isActive {
		state = 2
	}
	github_webhook, gw_found := triggers["github_webhook"]
	if !gw_found {
		github_webhook = false
	}
	schedule, s_found := triggers["schedule"]
	if !s_found {
		schedule = false
	}
	custom_branch_only, cbo_found := triggers["custom_branch_only"]
	if !cbo_found {
		custom_branch_only = false
	}
	jobTriggers := JobTrigger{
		Github_Webhook:     github_webhook.(bool),
		Schedule:           schedule.(bool),
		Custom_Branch_Only: custom_branch_only.(bool),
	}
	jobSettings := JobSettings{
		Threads:     numThreads,
		Target_Name: targetName,
	}
	jobSchedule := JobSchedule{
		Date: scheduleDate{
			Type: "every_day",
		},
		Time: scheduleTime{
			Type:     "every_hour",
			Interval: 1,
		},
	}

	newJob := Job{
		Account_Id:           c.AccountID,
		Project_Id:           projectId,
		Environment_Id:       environmentId,
		Name:                 name,
		Execute_Steps:        executeSteps,
		State:                state,
		Triggers:             jobTriggers,
		Settings:             jobSettings,
		Schedule:             jobSchedule,
		Generate_Docs:        generateDocs,
		Run_Generate_Sources: runGenerateSources,
	}
	if dbtVersion != "" {
		newJob.Dbt_Version = &dbtVersion
	}
	newJobData, err := json.Marshal(newJob)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/accounts/%s/v2/jobs/", c.HostURL, c.AccountID), strings.NewReader(string(newJobData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	jobResponse := JobResponse{}
	err = json.Unmarshal(body, &jobResponse)
	if err != nil {
		return nil, err
	}

	return &jobResponse.Data, nil
}

func (c *Client) UpdateJob(jobId string, job Job) (*Job, error) {
	jobData, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/accounts/%s/v2/%s/jobs/%s/", c.HostURL, c.AccountID, jobId), strings.NewReader(string(jobData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	jobResponse := JobResponse{}
	err = json.Unmarshal(body, &jobResponse)
	if err != nil {
		return nil, err
	}

	return &jobResponse.Data, nil
}
