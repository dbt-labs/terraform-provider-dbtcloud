package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type responseStatus struct {
	Code              int    `json:"code"`
	Is_Success        bool   `json:"is_success"`
	User_Message      string `json:"user_message"`
	Developer_Message string `json:"developer_message"`
}

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

type JobData struct {
	Account_Id     int         `json:"account_id"`
	Project_Id     int         `json:"project_id"`
	Environment_Id int         `json:"environment_id"`
	Name           string      `json:"name"`
	Execute_Steps  []string    `json:"execute_steps"`
	Dbt_Version    string      `json:"dbt_version"`
	Triggers       JobTrigger  `json:"triggers"`
	Settings       JobSettings `json:"settings"`
	State          int         `json:"state"`
	Generate_Docs  bool        `json:"generate_docs"`
	Schedule       JobSchedule `json:"schedule"`
}

type JobResponse struct {
	Data   JobData        `json:"data"`
	Status responseStatus `json:"status"`
}

type Job struct {
	ID             int         `json:"id,omitempty"`
	Account_Id     int         `json:"account_id"`
	Project_Id     int         `json:"project_id"`
	Environment_Id int         `json:"environment_id"`
	Name           string      `json:"name"`
	Execute_Steps  []string    `json:"execute_steps"`
	Dbt_Version    string      `json:"dbt_version"`
	Triggers       JobTrigger  `json:"triggers"`
	Settings       JobSettings `json:"settings"`
	State          int         `json:"state"`
	Generate_Docs  bool        `json:"generate_docs"`
	Schedule       JobSchedule `json:"schedule"`
}

func (c *Client) GetJob(jobID string) (*Job, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/jobs/%s", c.AccountURL, jobID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	job := Job{}
	err = json.Unmarshal(body, &job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (c *Client) CreateJob(projectId int, environmentId int, name string) (*Job, error) {
	newJob := JobData{
		Project_Id:     projectId,
		Environment_Id: environmentId,
		Name:           name,
	}
	newJobData, err := json.Marshal(newJob)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/jobs", c.AccountURL), strings.NewReader(string(newJobData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	job := Job{}
	err = json.Unmarshal(body, &job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}
