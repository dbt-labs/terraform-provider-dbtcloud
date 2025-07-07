package repository

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RepositoryDataSourceModel struct {
	ID                                    types.String `tfsdk:"id"`
	RepositoryID                          types.Int64  `tfsdk:"repository_id"`
	ProjectID                             types.Int64  `tfsdk:"project_id"`
	IsActive                              types.Bool   `tfsdk:"is_active"`
	RemoteURL                             types.String `tfsdk:"remote_url"`
	GitCloneStrategy                      types.String `tfsdk:"git_clone_strategy"`
	RepositoryCredentialsID               types.Int64  `tfsdk:"repository_credentials_id"`
	GitlabProjectID                       types.Int64  `tfsdk:"gitlab_project_id"`
	GithubInstallationID                  types.Int64  `tfsdk:"github_installation_id"`
	PrivateLinkEndpointID                 types.String `tfsdk:"private_link_endpoint_id"`
	DeployKey                             types.String `tfsdk:"deploy_key"`
	PullRequestURLTemplate                types.String `tfsdk:"pull_request_url_template"`
	AzureActiveDirectoryProjectID         types.String `tfsdk:"azure_active_directory_project_id"`
	AzureActiveDirectoryRepositoryID      types.String `tfsdk:"azure_active_directory_repository_id"`
	AzureBypassWebhookRegistrationFailure types.Bool   `tfsdk:"azure_bypass_webhook_registration_failure"`
	FetchDeployKey                        types.Bool   `tfsdk:"fetch_deploy_key"`
}

type RepositoryResourceModel struct {
	ID                                    types.String `tfsdk:"id"`
	RepositoryID                          types.Int64  `tfsdk:"repository_id"`
	ProjectID                             types.Int64  `tfsdk:"project_id"`
	IsActive                              types.Bool   `tfsdk:"is_active"`
	RemoteURL                             types.String `tfsdk:"remote_url"`
	GitCloneStrategy                      types.String `tfsdk:"git_clone_strategy"`
	RepositoryCredentialsID               types.Int64  `tfsdk:"repository_credentials_id"`
	GitlabProjectID                       types.Int64  `tfsdk:"gitlab_project_id"`
	GithubInstallationID                  types.Int64  `tfsdk:"github_installation_id"`
	PrivateLinkEndpointID                 types.String `tfsdk:"private_link_endpoint_id"`
	DeployKey                             types.String `tfsdk:"deploy_key"`
	PullRequestURLTemplate                types.String `tfsdk:"pull_request_url_template"`
	AzureActiveDirectoryProjectID         types.String `tfsdk:"azure_active_directory_project_id"`
	AzureActiveDirectoryRepositoryID      types.String `tfsdk:"azure_active_directory_repository_id"`
	AzureBypassWebhookRegistrationFailure types.Bool   `tfsdk:"azure_bypass_webhook_registration_failure"`
	FetchDeployKey                        types.Bool   `tfsdk:"fetch_deploy_key"`
}
