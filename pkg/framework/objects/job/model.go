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
	Execution                     JobExecution          `tfsdk:"execution"`
	GenerateDocs                  types.Bool            `tfsdk:"generate_docs"`
	RunGenerateSources            types.Bool            `tfsdk:"run_generate_sources"`
	ID                            types.Int64           `tfsdk:"id"`
	ProjectID                     types.Int64           `tfsdk:"project_id"`
	EnvironmentID                 types.Int64           `tfsdk:"environment_id"`
	Name                          types.String          `tfsdk:"name"`
	Description                   types.String          `tfsdk:"description"`
	DbtVersion                    types.String          `tfsdk:"dbt_version"`
	ExecuteSteps                  []types.String        `tfsdk:"execute_steps"`
	DeferringJobDefinitionID      types.Int64           `tfsdk:"deferring_job_definition_id"`
	DeferringEnvironmentID        types.Int64           `tfsdk:"deferring_environment_id"`
	Triggers                      JobTriggers           `tfsdk:"triggers"`
	Settings                      JobSettings           `tfsdk:"settings"`
	Schedule                      JobSchedule           `tfsdk:"schedule"`
	JobType                       types.String          `tfsdk:"job_type"`
	TriggersOnDraftPr             types.Bool            `tfsdk:"triggers_on_draft_pr"`
	Environment                   JobEnvironment        `tfsdk:"environment"`
	JobCompletionTriggerCondition *JobCompletionTrigger `tfsdk:"job_completion_trigger_condition"`
}
