package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type JobTrigger struct {
	Github_Webhook     bool `json:"github_webhook"`
	Schedule           bool `json:"schedule"`
	Custom_Branch_Only bool `json:"custom_branch_only"`
	GitProviderWebhook bool `json:"git_provider_webhook"`
}

type JobSettings struct {
	Threads     int    `json:"threads"`
	Target_Name string `json:"target_name"`
}

type scheduleDate struct {
	Type string  `json:"type"`
	Days *[]int  `json:"days,omitempty"`
	Cron *string `json:"cron,omitempty"`
}

type scheduleTime struct {
	Type     string `json:"type"`
	Interval int    `json:"interval,omitempty"`
	Hours    *[]int `json:"hours,omitempty"`
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
	Deferring_Job_Id     *int        `json:"deferring_job_definition_id"`
}

func (c *Client) GetJob(jobID string) (*Job, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v2/accounts/%s/jobs/%s/", c.HostURL, strconv.Itoa(c.AccountID), jobID), nil)
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

func (c *Client) CreateJob(projectId int, environmentId int, name string, executeSteps []string, dbtVersion string, isActive bool, triggers map[string]interface{}, numThreads int, targetName string, generateDocs bool, runGenerateSources bool, scheduleType string, scheduleInterval int, scheduleHours []int, scheduleDays []int, scheduleCron string, deferringJobId int) (*Job, error) {
	state := STATE_ACTIVE
	if !isActive {
		state = STATE_DELETED
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
	git_provider_webhook, gpw_found := triggers["git_provider_webhook"]
	if !gpw_found {
		git_provider_webhook = false
	}
	jobTriggers := JobTrigger{
		Github_Webhook:     github_webhook.(bool),
		Schedule:           schedule.(bool),
		Custom_Branch_Only: custom_branch_only.(bool),
		GitProviderWebhook: git_provider_webhook.(bool),
	}
	jobSettings := JobSettings{
		Threads:     numThreads,
		Target_Name: targetName,
	}

	time := scheduleTime{
		Type:     "every_hour",
		Interval: 1,
	}
	if scheduleInterval > 0 {
		time.Interval = scheduleInterval
	}
	if len(scheduleHours) > 0 {
		time.Type = "at_exact_hours"
		time.Hours = &scheduleHours
		time.Interval = 0
	}

	date := scheduleDate{
		Type: scheduleType,
	}
	if scheduleType == "days_of_week" {
		date.Days = &scheduleDays
	} else if scheduleCron != "" {
		date.Cron = &scheduleCron
	}
	jobSchedule := JobSchedule{
		Date: date,
		Time: time,
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
	if deferringJobId != 0 {
		newJob.Deferring_Job_Id = &deferringJobId
	}
	newJobData, err := json.Marshal(newJob)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/accounts/%s/jobs/", c.HostURL, strconv.Itoa(c.AccountID)), strings.NewReader(string(newJobData)))
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

	fmt.Printf(string(jobData))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/accounts/%s/jobs/%s/", c.HostURL, strconv.Itoa(c.AccountID), jobId), strings.NewReader(string(jobData)))
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
