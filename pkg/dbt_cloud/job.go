package dbt_cloud

type ResponseStatus struct {
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

type Job struct {
	Data   JobData        `json:"data"`
	Status ResponseStatus `json:"status"`
}
