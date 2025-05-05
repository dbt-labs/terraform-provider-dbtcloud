package runs

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RunDataSourceModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	AccountID           types.Int64  `tfsdk:"account_id"`
	JobID               types.Int64  `tfsdk:"job_id"`
	GitSHA              types.String `tfsdk:"git_sha"`
	GitBranch           types.String `tfsdk:"git_branch"`
	GitHubPullRequestID types.String `tfsdk:"github_pull_request_id"`
	SchemaOverride      types.String `tfsdk:"schema_override"`
	Cause               types.String `tfsdk:"cause"`
}

type RunFilterModel struct {
	EnvironmentID   types.Int64  `tfsdk:"environment_id"`
	Limit           types.Int64  `tfsdk:"limit"`
	ProjectID       types.Int64  `tfsdk:"project_id"`
	TriggerID       types.Int64  `tfsdk:"trigger_id"`
	JobDefinitionID types.Int64  `tfsdk:"job_definition_id"`
	PullRequestID   types.Int64  `tfsdk:"pull_request_id"`
	Status          types.Int64  `tfsdk:"status"`
	StatusIn        types.String `tfsdk:"status_in"`
}

type RunsDataSourceModel struct {
	Filter RunFilterModel       `tfsdk:"filter"`
	Runs   []RunDataSourceModel `tfsdk:"runs"`
}
