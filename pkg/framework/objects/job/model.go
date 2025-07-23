package job

import "github.com/hashicorp/terraform-plugin-framework/types"

type JobsDataSourceModel struct {
	ProjectID     types.Int64          `tfsdk:"project_id"`
	EnvironmentID types.Int64          `tfsdk:"environment_id"`
	Jobs          []JobDataSourceModel `tfsdk:"jobs"`
}

type JobExecution struct {
	TimeoutSeconds types.Int64 `tfsdk:"timeout_seconds"`
}

type JobTriggers struct {
	GithubWebhook      types.Bool `tfsdk:"github_webhook"`
	GitProviderWebhook types.Bool `tfsdk:"git_provider_webhook"`
	Schedule           types.Bool `tfsdk:"schedule"`
	OnMerge            types.Bool `tfsdk:"on_merge"`
}

type JobSettings struct {
	Threads    types.Int64  `tfsdk:"threads"`
	TargetName types.String `tfsdk:"target_name"`
}

type JobEnvironment struct {
	ProjectID      types.Int64  `tfsdk:"project_id"`
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	DeploymentType types.String `tfsdk:"deployment_type"`
	Type           types.String `tfsdk:"type"`
}

type JobCompletionTrigger struct {
	Condition JobCompletionTriggerCondition `tfsdk:"condition"`
}

type JobCompletionTriggerCondition struct {
	JobID     types.Int64    `tfsdk:"job_id"`
	ProjectID types.Int64    `tfsdk:"project_id"`
	Statuses  []types.String `tfsdk:"statuses"`
}

type JobSchedule struct {
	Cron types.String `tfsdk:"cron"`
}

type JobDataSourceModel struct {
	Execution                     *JobExecution         `tfsdk:"execution"`
	TimeoutSeconds                types.Int64           `tfsdk:"timeout_seconds"`
	GenerateDocs                  types.Bool            `tfsdk:"generate_docs"`
	RunGenerateSources            types.Bool            `tfsdk:"run_generate_sources"`
	ID                            types.Int64           `tfsdk:"id"`
	JobId                         types.Int64           `tfsdk:"job_id"`
	ProjectID                     types.Int64           `tfsdk:"project_id"`
	EnvironmentID                 types.Int64           `tfsdk:"environment_id"`
	Name                          types.String          `tfsdk:"name"`
	Description                   types.String          `tfsdk:"description"`
	DbtVersion                    types.String          `tfsdk:"dbt_version"`
	ExecuteSteps                  []types.String        `tfsdk:"execute_steps"`
	DeferringJobDefinitionID      types.Int64           `tfsdk:"deferring_job_definition_id"`
	DeferringEnvironmentID        types.Int64           `tfsdk:"deferring_environment_id"`
	Triggers                      *JobTriggers          `tfsdk:"triggers"`
	Settings                      *JobSettings          `tfsdk:"settings"`
	Schedule                      *JobSchedule          `tfsdk:"schedule"`
	JobType                       types.String          `tfsdk:"job_type"`
	TriggersOnDraftPr             types.Bool            `tfsdk:"triggers_on_draft_pr"`
	Environment                   *JobEnvironment       `tfsdk:"environment"`
	JobCompletionTriggerCondition *JobCompletionTrigger `tfsdk:"job_completion_trigger_condition"`
	RunCompareChanges             types.Bool            `tfsdk:"run_compare_changes"`
}

// TODO remove this in the next major release
type SingleJobDataSourceModel struct {
	Execution                     *JobExecution                    `tfsdk:"execution"`
	TimeoutSeconds                types.Int64                      `tfsdk:"timeout_seconds"`
	GenerateDocs                  types.Bool                       `tfsdk:"generate_docs"`
	RunGenerateSources            types.Bool                       `tfsdk:"run_generate_sources"`
	ID                            types.Int64                      `tfsdk:"id"`
	JobId                         types.Int64                      `tfsdk:"job_id"`
	ProjectID                     types.Int64                      `tfsdk:"project_id"`
	EnvironmentID                 types.Int64                      `tfsdk:"environment_id"`
	Name                          types.String                     `tfsdk:"name"`
	Description                   types.String                     `tfsdk:"description"`
	DbtVersion                    types.String                     `tfsdk:"dbt_version"`
	ExecuteSteps                  []types.String                   `tfsdk:"execute_steps"`
	DeferringJobId                types.Int64                      `tfsdk:"deferring_job_id"`
	DeferringEnvironmentID        types.Int64                      `tfsdk:"deferring_environment_id"`
	SelfDeferring                 types.Bool                       `tfsdk:"self_deferring"`
	Triggers                      *JobTriggers                     `tfsdk:"triggers"`
	Settings                      *JobSettings                     `tfsdk:"settings"`
	Schedule                      *JobSchedule                     `tfsdk:"schedule"`
	JobType                       types.String                     `tfsdk:"job_type"`
	TriggersOnDraftPr             types.Bool                       `tfsdk:"triggers_on_draft_pr"`
	Environment                   *JobEnvironment                  `tfsdk:"environment"`
	JobCompletionTriggerCondition []*JobCompletionTriggerCondition `tfsdk:"job_completion_trigger_condition"`
	RunCompareChanges             types.Bool                       `tfsdk:"run_compare_changes"`
}

type JobResourceModel struct {
	Execution              *JobExecution  `tfsdk:"execution"`                // has timeout-seconds
	GenerateDocs           types.Bool     `tfsdk:"generate_docs"`            // exists
	RunGenerateSources     types.Bool     `tfsdk:"run_generate_sources"`     // exists
	ID                     types.Int64    `tfsdk:"id"`                       // will hold job id?
	JobId                  types.Int64    `tfsdk:"job_id"`                   // for framework
	ProjectID              types.Int64    `tfsdk:"project_id"`               // exists
	EnvironmentID          types.Int64    `tfsdk:"environment_id"`           // exists
	Name                   types.String   `tfsdk:"name"`                     // exists
	Description            types.String   `tfsdk:"description"`              // exists
	DbtVersion             types.String   `tfsdk:"dbt_version"`              // exists
	ExecuteSteps           []types.String `tfsdk:"execute_steps"`            // exists
	DeferringEnvironmentID types.Int64    `tfsdk:"deferring_environment_id"` // exists
	Triggers               *JobTriggers   `tfsdk:"triggers"`                 // exists
	// Settings                      *JobSettings          `tfsdk:"settings"`                 // has no of threads and target name
	// Schedule                      *JobSchedule          `tfsdk:"schedule"`                 // has cron expression
	JobType           types.String `tfsdk:"job_type"`             // exists
	TriggersOnDraftPr types.Bool   `tfsdk:"triggers_on_draft_pr"` // exists
	// Environment                   *JobEnvironment       `tfsdk:"environment"`
	JobCompletionTriggerCondition []*JobCompletionTriggerCondition `tfsdk:"job_completion_trigger_condition"` // exists
	RunCompareChanges             types.Bool                       `tfsdk:"run_compare_changes"`              // exists
	IsActive                      types.Bool                       `tfsdk:"is_active"`
	TargetName                    types.String                     `tfsdk:"target_name"` // add deprecated
	NumThreads                    types.Int64                      `tfsdk:"num_threads"` // add deprecated moved to settings
	RunLint                       types.Bool                       `tfsdk:"run_lint"`
	ErrorsOnLintFailure           types.Bool                       `tfsdk:"errors_on_lint_failure"`
	ScheduleType                  types.String                     `tfsdk:"schedule_type"`
	ScheduleInterval              types.Int64                      `tfsdk:"schedule_interval"`
	ScheduleHours                 []types.Int64                    `tfsdk:"schedule_hours"`
	ScheduleDays                  []types.Int64                    `tfsdk:"schedule_days"`
	ScheduleCron                  types.String                     `tfsdk:"schedule_cron"`    // add deprecated move to schedule
	DeferringJobId                types.Int64                      `tfsdk:"deferring_job_id"` // add deprecated move to deferring_job_definition_id
	SelfDeferring                 types.Bool                       `tfsdk:"self_deferring"`
	CompareChangesFlags           types.String                     `tfsdk:"compare_changes_flags"`
}
