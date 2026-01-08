package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &repositoryResource{}
	_ resource.ResourceWithConfigure   = &repositoryResource{}
	_ resource.ResourceWithImportState = &repositoryResource{}
)

func RepositoryResource() resource.Resource {
	return &repositoryResource{}
}

type repositoryResource struct {
	client *dbt_cloud.Client
}

func (r *repositoryResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *repositoryResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = ResourceSchema()
}

func (r *repositoryResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan RepositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	remoteURL := plan.RemoteURL.ValueString()
	isActive := plan.IsActive.ValueBool()
	gitCloneStrategy := plan.GitCloneStrategy.ValueString()
	pullRequestURLTemplate := plan.PullRequestURLTemplate.ValueString()

	var gitlabProjectID int
	var githubInstallationID int
	var privateLinkEndpointID string
	var azureProjectID string
	var azureRepositoryID string
	var azureBypassWebhookRegistrationFailure bool

	if !plan.GitlabProjectID.IsNull() {
		gitlabProjectID = int(plan.GitlabProjectID.ValueInt64())
	}

	if !plan.GithubInstallationID.IsNull() {
		githubInstallationID = int(plan.GithubInstallationID.ValueInt64())
	}

	if !plan.PrivateLinkEndpointID.IsNull() {
		privateLinkEndpointID = plan.PrivateLinkEndpointID.ValueString()
	}

	if !plan.AzureActiveDirectoryProjectID.IsNull() {
		azureProjectID = plan.AzureActiveDirectoryProjectID.ValueString()
	}

	if !plan.AzureActiveDirectoryRepositoryID.IsNull() {
		azureRepositoryID = plan.AzureActiveDirectoryRepositoryID.ValueString()
	}

	if !plan.AzureBypassWebhookRegistrationFailure.IsNull() {
		azureBypassWebhookRegistrationFailure = plan.AzureBypassWebhookRegistrationFailure.ValueBool()
	}

	repository, err := r.client.CreateRepository(
		projectID,
		remoteURL,
		isActive,
		gitCloneStrategy,
		gitlabProjectID,
		githubInstallationID,
		privateLinkEndpointID,
		azureProjectID,
		azureRepositoryID,
		azureBypassWebhookRegistrationFailure,
		pullRequestURLTemplate,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating repository",
			err.Error(),
		)
		return
	}

	// checking potential issues with the creation of GitLab repositories with service tokens
	if repository.RepositoryCredentialsID == nil && gitlabProjectID != 0 {
		repositoryIDString := fmt.Sprintf("%d", *repository.ID)
		projectIDString := fmt.Sprintf("%d", repository.ProjectID)
		_, err := r.client.DeleteRepository(repositoryIDString, projectIDString)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error deleting invalid repository",
				err.Error(),
			)
			return
		}

		resp.Diagnostics.AddError(
			"Invalid repository configuration",
			"`repository_credentials_id` is not set after creating the repository. This is likely due to creating the repository with a service token. Only user tokens / personal access tokens are supported for GitLab at the moment",
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d%s%d", repository.ProjectID, dbt_cloud.ID_DELIMITER, *repository.ID))
	plan.RepositoryID = types.Int64Value(int64(*repository.ID))
	plan.IsActive = types.BoolValue(repository.State == dbt_cloud.STATE_ACTIVE)
	plan.ProjectID = types.Int64Value(int64(repository.ProjectID))
	plan.RemoteURL = types.StringValue(repository.RemoteUrl)
	// When github_installation_id is provided, the API automatically changes git_clone_strategy to "github_app"
	// We must use the API's returned value to avoid Terraform detecting a mismatch and wanting to replace the resource
	if githubInstallationID != 0 {
		// github_installation_id was provided, use API's returned value (will be "github_app")
		plan.GitCloneStrategy = types.StringValue(repository.GitCloneStrategy)
	} else {
		// No github_installation_id, preserve planned value to handle case differences, etc.
		// plan.GitCloneStrategy is already set from the plan, don't overwrite it
	}

	if repository.RepositoryCredentialsID != nil {
		plan.RepositoryCredentialsID = types.Int64Value(int64(*repository.RepositoryCredentialsID))
	} else {
		plan.RepositoryCredentialsID = types.Int64Null()
	}

	if repository.GitlabProjectID != nil {
		plan.GitlabProjectID = types.Int64Value(int64(*repository.GitlabProjectID))
	} else if gitlabProjectID != 0 {
		plan.GitlabProjectID = types.Int64Value(int64(gitlabProjectID))
	} else {
		plan.GitlabProjectID = types.Int64Null()
	}

	if repository.GithubInstallationID != nil {
		plan.GithubInstallationID = types.Int64Value(int64(*repository.GithubInstallationID))
	} else if githubInstallationID != 0 {
		plan.GithubInstallationID = types.Int64Value(int64(githubInstallationID))
	} else {
		plan.GithubInstallationID = types.Int64Null()
	}

	if repository.PrivateLinkEndpointID != nil {
		plan.PrivateLinkEndpointID = types.StringValue(*repository.PrivateLinkEndpointID)
	} else if privateLinkEndpointID != "" {
		plan.PrivateLinkEndpointID = types.StringValue(privateLinkEndpointID)
	} else {
		plan.PrivateLinkEndpointID = types.StringNull()
	}

	if repository.DeployKey != nil {
		plan.DeployKey = types.StringValue(repository.DeployKey.PublicKey)
	} else {
		plan.DeployKey = types.StringNull()
	}

	if repository.PullRequestURLTemplate != "" {
		plan.PullRequestURLTemplate = types.StringValue(repository.PullRequestURLTemplate)
	} else if pullRequestURLTemplate != "" {
		plan.PullRequestURLTemplate = types.StringValue(pullRequestURLTemplate)
	} else {
		plan.PullRequestURLTemplate = types.StringNull()
	}

	if repository.AzureActiveDirectoryProjectID != nil {
		plan.AzureActiveDirectoryProjectID = types.StringValue(*repository.AzureActiveDirectoryProjectID)
	} else if azureProjectID != "" {
		plan.AzureActiveDirectoryProjectID = types.StringValue(azureProjectID)
	} else {
		plan.AzureActiveDirectoryProjectID = types.StringValue("")
	}

	if repository.AzureActiveDirectoryRepositoryID != nil {
		plan.AzureActiveDirectoryRepositoryID = types.StringValue(*repository.AzureActiveDirectoryRepositoryID)
	} else if azureRepositoryID != "" {
		plan.AzureActiveDirectoryRepositoryID = types.StringValue(azureRepositoryID)
	} else {
		plan.AzureActiveDirectoryRepositoryID = types.StringValue("")
	}

	if repository.AzureBypassWebhookRegistrationFailure != nil {
		plan.AzureBypassWebhookRegistrationFailure = types.BoolValue(*repository.AzureBypassWebhookRegistrationFailure)
	} else {
		plan.AzureBypassWebhookRegistrationFailure = types.BoolValue(azureBypassWebhookRegistrationFailure)
	}

	// Handle the deprecated FetchDeployKey field - just maintain the planned value
	// This field doesn't affect API behavior but needs to be consistent for Terraform
	if plan.FetchDeployKey.IsNull() {
		plan.FetchDeployKey = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *repositoryResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state RepositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(state.ID.ValueString(), dbt_cloud.ID_DELIMITER)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid ID format",
			fmt.Sprintf("Expected ID in format project_id%srepository_id, got: %s", dbt_cloud.ID_DELIMITER, state.ID.ValueString()),
		)
		return
	}

	projectID := parts[0]
	repositoryID := parts[1]

	repository, err := r.client.GetRepository(repositoryID, projectID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading repository",
			err.Error(),
		)
		return
	}

	state.IsActive = types.BoolValue(repository.State == dbt_cloud.STATE_ACTIVE)
	state.ProjectID = types.Int64Value(int64(repository.ProjectID))
	state.RepositoryID = types.Int64Value(int64(*repository.ID))
	state.RemoteURL = types.StringValue(repository.RemoteUrl)
	state.GitCloneStrategy = types.StringValue(repository.GitCloneStrategy)

	if repository.RepositoryCredentialsID != nil {
		state.RepositoryCredentialsID = types.Int64Value(int64(*repository.RepositoryCredentialsID))
	} else {
		state.RepositoryCredentialsID = types.Int64Null()
	}

	if repository.GitlabProjectID != nil { // GitlabProjectID is not returned by the api for the moment, it always return null
		state.GitlabProjectID = types.Int64Value(int64(*repository.GitlabProjectID))
	}

	if repository.GithubInstallationID != nil {
		state.GithubInstallationID = types.Int64Value(int64(*repository.GithubInstallationID))
	} else {
		state.GithubInstallationID = types.Int64Null()
	}

	if repository.PrivateLinkEndpointID != nil {
		state.PrivateLinkEndpointID = types.StringValue(*repository.PrivateLinkEndpointID)
	} else {
		state.PrivateLinkEndpointID = types.StringNull()
	}

	if repository.DeployKey != nil {
		state.DeployKey = types.StringValue(repository.DeployKey.PublicKey)
	} else {
		state.DeployKey = types.StringNull()
	}

	if repository.PullRequestURLTemplate != "" {
		state.PullRequestURLTemplate = types.StringValue(repository.PullRequestURLTemplate)
	} else {
		state.PullRequestURLTemplate = types.StringNull()
	}

	/// AAD Project ID and Repo ID always come up as nil from the API response
	if repository.AzureActiveDirectoryProjectID != nil {
		state.AzureActiveDirectoryProjectID = types.StringValue(*repository.AzureActiveDirectoryProjectID)
	} else if state.AzureActiveDirectoryProjectID.IsNull() {
		state.AzureActiveDirectoryProjectID = types.StringValue("")
	}

	if repository.AzureActiveDirectoryRepositoryID != nil {
		state.AzureActiveDirectoryRepositoryID = types.StringValue(*repository.AzureActiveDirectoryRepositoryID)
	} else if state.AzureActiveDirectoryRepositoryID.IsNull() {
		state.AzureActiveDirectoryRepositoryID = types.StringValue("")
	}

	if repository.AzureBypassWebhookRegistrationFailure != nil {
		state.AzureBypassWebhookRegistrationFailure = types.BoolValue(*repository.AzureBypassWebhookRegistrationFailure)
	} else {
		state.AzureBypassWebhookRegistrationFailure = types.BoolValue(false)
	}

	// Handle the deprecated FetchDeployKey field - maintain existing value or default to false
	// This field doesn't affect API behavior but needs to be consistent for Terraform
	if state.FetchDeployKey.IsNull() {
		state.FetchDeployKey = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *repositoryResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan RepositoryResourceModel
	var state RepositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasIsActiveChange := !plan.IsActive.Equal(state.IsActive)
	hasPullRequestURLTemplateChange := !plan.PullRequestURLTemplate.Equal(state.PullRequestURLTemplate)

	parts := strings.Split(state.ID.ValueString(), dbt_cloud.ID_DELIMITER)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid ID format",
			fmt.Sprintf("Expected ID in format project_id%srepository_id, got: %s", dbt_cloud.ID_DELIMITER, state.ID.ValueString()),
		)
		return
	}
	projectID := parts[0]
	repositoryID := parts[1]

	repository, err := r.client.GetRepository(repositoryID, projectID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading repository",
			err.Error(),
		)
		return
	}

	if hasIsActiveChange {
		if plan.IsActive.ValueBool() {
			repository.State = dbt_cloud.STATE_ACTIVE
		} else {
			repository.State = dbt_cloud.STATE_DELETED
		}
	}

	if hasPullRequestURLTemplateChange {
		repository.PullRequestURLTemplate = plan.PullRequestURLTemplate.ValueString()
	}

	updatedRepository, err := r.client.UpdateRepository(repositoryID, projectID, *repository)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating repository",
			err.Error(),
		)
		return
	}

	state.IsActive = types.BoolValue(updatedRepository.State == dbt_cloud.STATE_ACTIVE)
	state.ProjectID = types.Int64Value(int64(updatedRepository.ProjectID))
	state.RepositoryID = types.Int64Value(int64(*updatedRepository.ID))
	state.RemoteURL = types.StringValue(updatedRepository.RemoteUrl)
	state.GitCloneStrategy = types.StringValue(updatedRepository.GitCloneStrategy)

	if updatedRepository.RepositoryCredentialsID != nil {
		state.RepositoryCredentialsID = types.Int64Value(int64(*updatedRepository.RepositoryCredentialsID))
	} else {
		state.RepositoryCredentialsID = types.Int64Null()
	}

	if updatedRepository.GitlabProjectID != nil {
		state.GitlabProjectID = types.Int64Value(int64(*updatedRepository.GitlabProjectID))
	}

	if updatedRepository.GithubInstallationID != nil {
		state.GithubInstallationID = types.Int64Value(int64(*updatedRepository.GithubInstallationID))
	} else {
		state.GithubInstallationID = types.Int64Null()
	}

	if updatedRepository.PrivateLinkEndpointID != nil {
		state.PrivateLinkEndpointID = types.StringValue(*updatedRepository.PrivateLinkEndpointID)
	} else {
		state.PrivateLinkEndpointID = types.StringNull()
	}

	if updatedRepository.DeployKey != nil {
		state.DeployKey = types.StringValue(updatedRepository.DeployKey.PublicKey)
	} else {
		state.DeployKey = types.StringNull()
	}

	if updatedRepository.PullRequestURLTemplate != "" {
		state.PullRequestURLTemplate = types.StringValue(updatedRepository.PullRequestURLTemplate)
	} else {
		state.PullRequestURLTemplate = types.StringNull()
	}

	if updatedRepository.AzureActiveDirectoryProjectID != nil {
		state.AzureActiveDirectoryProjectID = types.StringValue(*updatedRepository.AzureActiveDirectoryProjectID)
	} else {
		state.AzureActiveDirectoryProjectID = types.StringValue("")
	}

	if updatedRepository.AzureActiveDirectoryRepositoryID != nil {
		state.AzureActiveDirectoryRepositoryID = types.StringValue(*updatedRepository.AzureActiveDirectoryRepositoryID)
	} else {
		state.AzureActiveDirectoryRepositoryID = types.StringValue("")
	}

	if updatedRepository.AzureBypassWebhookRegistrationFailure != nil {
		state.AzureBypassWebhookRegistrationFailure = types.BoolValue(*updatedRepository.AzureBypassWebhookRegistrationFailure)
	} else {
		state.AzureBypassWebhookRegistrationFailure = types.BoolValue(false)
	}

	// Handle the deprecated FetchDeployKey field - maintain the planned value
	// This field doesn't affect API behavior but needs to be consistent for Terraform
	state.FetchDeployKey = plan.FetchDeployKey

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *repositoryResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state RepositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(state.ID.ValueString(), dbt_cloud.ID_DELIMITER)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid ID format",
			fmt.Sprintf("Expected ID in format project_id%srepository_id, got: %s", dbt_cloud.ID_DELIMITER, state.ID.ValueString()),
		)
		return
	}
	projectID := parts[0]
	repositoryID := parts[1]

	_, err := r.client.DeleteRepository(repositoryID, projectID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting repository",
			err.Error(),
		)
		return
	}
}

func (r *repositoryResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.Split(req.ID, dbt_cloud.ID_DELIMITER)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format project_id%srepository_id, got: %s", dbt_cloud.ID_DELIMITER, req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *repositoryResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
