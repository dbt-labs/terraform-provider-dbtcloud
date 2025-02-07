package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type JobTrigger struct {
	GithubWebhook      bool `json:"github_webhook"`
	Schedule           bool `json:"schedule"`
	GitProviderWebhook bool `json:"git_provider_webhook"`
	OnMerge            bool `json:"on_merge"`
}

type JobSettings struct {
	Threads    int    `json:"threads"`
	TargetName string `json:"target_name"`
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

type JobExecution struct {
	TimeoutSeconds int `json:"timeout_seconds"`
}

type JobCompletionTrigger struct {
	Condition JobCompletionTriggerCondition `json:"condition"`
}

type JobCompletionTriggerCondition struct {
	JobID     int   `json:"job_id"`
	ProjectID int   `json:"project_id"`
	Statuses  []int `json:"statuses"`
}

type Job struct {
	ID                     *int                  `json:"id"`
	AccountId              int                   `json:"account_id"`
	ProjectId              int                   `json:"project_id"`
	EnvironmentId          int                   `json:"environment_id"`
	Name                   string                `json:"name"`
	CompareChangesFlags    string                `json:"compare_changes_flags"`
	DbtVersion             *string               `json:"dbt_version"`
	DeferringEnvironmentId *int                  `json:"deferring_environment_id"`
	DeferringJobId         *int                  `json:"deferring_job_definition_id"`
	Description            string                `json:"description"`
	ErrorsOnLintFailure    bool                  `json:"errors_on_lint_failure"`
	ExecuteSteps           []string              `json:"execute_steps"`
	Execution              JobExecution          `json:"execution"`
	GenerateDocs           bool                  `json:"generate_docs"`
	JobCompletionTrigger   *JobCompletionTrigger `json:"job_completion_trigger_condition"`
	JobType                string                `json:"job_type,omitempty"`
	RunCompareChanges      bool                  `json:"run_compare_changes"`
	RunGenerateSources     bool                  `json:"run_generate_sources"`
	RunLint                bool                  `json:"run_lint"`
	Schedule               JobSchedule           `json:"schedule"`
	Settings               JobSettings           `json:"settings"`
	State                  int                   `json:"state"`
	TriggersOnDraftPR      bool                  `json:"triggers_on_draft_pr"`
	Triggers               JobTrigger            `json:"triggers"`
}

type JobWithEnvironment struct {
	Job
	Environment Environment `json:"environment"`
}

func (c *Client) GetJob(jobID string) (*Job, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/v2/accounts/%s/jobs/%s/", c.HostURL, strconv.Itoa(c.AccountID), jobID),
		nil,
	)
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

func (c *Client) CreateJob(
	projectId int,
	environmentId int,
	name string,
	description string,
	executeSteps []string,
	dbtVersion string,
	isActive bool,
	triggers map[string]any,
	numThreads int,
	targetName string,
	generateDocs bool,
	runGenerateSources bool,
	scheduleType string,
	scheduleInterval int,
	scheduleHours []int,
	scheduleDays []int,
	scheduleCron string,
	deferringJobId int,
	deferringEnvironmentID int,
	selfDeferring bool,
	timeoutSeconds int,
	triggersOnDraftPR bool,
	jobCompletionTriggerCondition map[string]any,
	runCompareChanges bool,
	runLint bool,
	errorsOnLintFailure bool,
	jobType string,
	compareChangesFlags string,
) (*Job, error) {
	state := STATE_ACTIVE
	if !isActive {
		state = STATE_DELETED
	}
	finalJobType := ""
	github_webhook, gw_found := triggers["github_webhook"]
	if !gw_found {
		github_webhook = false
	}
	schedule, s_found := triggers["schedule"]
	if !s_found {
		schedule = false
	}
	onMerge, s_found := triggers["on_merge"]
	if !s_found {
		onMerge = false
	}
	if onMerge.(bool) {
		finalJobType = "merge"
	}
	git_provider_webhook, gpw_found := triggers["git_provider_webhook"]
	if !gpw_found {
		git_provider_webhook = false
	}
	if git_provider_webhook.(bool) {
		finalJobType = "ci"
	}
	// we default to the provided job type if it is set
	if jobType != "" {
		finalJobType = jobType
	}
	jobTriggers := JobTrigger{
		GithubWebhook:      github_webhook.(bool),
		Schedule:           schedule.(bool),
		GitProviderWebhook: git_provider_webhook.(bool),
		OnMerge:            onMerge.(bool),
	}
	jobSettings := JobSettings{
		Threads:    numThreads,
		TargetName: targetName,
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
		date.Cron = nil
	} else if scheduleCron != "" {
		date.Cron = &scheduleCron
	}
	jobSchedule := JobSchedule{
		Date: date,
		Time: time,
	}
	jobExecution := JobExecution{
		TimeoutSeconds: timeoutSeconds,
	}

	jobCompletionTrigger := &JobCompletionTrigger{}
	if len(jobCompletionTriggerCondition) == 0 {
		jobCompletionTrigger = nil
	} else {
		jobCompletionTrigger = &JobCompletionTrigger{
			Condition: JobCompletionTriggerCondition{
				JobID:     jobCompletionTriggerCondition["job_id"].(int),
				ProjectID: jobCompletionTriggerCondition["project_id"].(int),
				Statuses:  jobCompletionTriggerCondition["statuses"].([]int),
			},
		}
	}

	newJob := Job{
		AccountId:            c.AccountID,
		ProjectId:            projectId,
		EnvironmentId:        environmentId,
		Name:                 name,
		Description:          description,
		ExecuteSteps:         executeSteps,
		State:                state,
		Triggers:             jobTriggers,
		Settings:             jobSettings,
		Schedule:             jobSchedule,
		GenerateDocs:         generateDocs,
		RunGenerateSources:   runGenerateSources,
		Execution:            jobExecution,
		TriggersOnDraftPR:    triggersOnDraftPR,
		JobCompletionTrigger: jobCompletionTrigger,
		JobType:              finalJobType,
		RunCompareChanges:    runCompareChanges,
		RunLint:              runLint,
		ErrorsOnLintFailure:  errorsOnLintFailure,
		CompareChangesFlags:  compareChangesFlags,
	}
	if dbtVersion != "" {
		newJob.DbtVersion = &dbtVersion
	}
	if deferringJobId != 0 {
		newJob.DeferringJobId = &deferringJobId
	} else {
		newJob.DeferringJobId = nil
	}
	if deferringEnvironmentID != 0 {
		newJob.DeferringEnvironmentId = &deferringEnvironmentID
	} else {
		newJob.DeferringEnvironmentId = nil
	}
	newJobData, err := json.Marshal(newJob)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/v2/accounts/%s/jobs/", c.HostURL, strconv.Itoa(c.AccountID)),
		strings.NewReader(string(newJobData)),
	)
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

	if selfDeferring {
		updatedJob := newJob
		deferringJobID := *jobResponse.Data.ID
		selfID := *jobResponse.Data.ID
		updatedJob.DeferringJobId = &deferringJobID
		updatedJob.ID = &selfID
		return c.UpdateJob(strconv.Itoa(*jobResponse.Data.ID), updatedJob)
	}

	return &jobResponse.Data, nil
}

func (c *Client) UpdateJob(jobId string, job Job) (*Job, error) {

	jobData, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/v2/accounts/%s/jobs/%s/", c.HostURL, strconv.Itoa(c.AccountID), jobId),
		strings.NewReader(string(jobData)),
	)
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
