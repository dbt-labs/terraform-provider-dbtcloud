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

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &repositoryResource{}
	_ resource.ResourceWithConfigure   = &repositoryResource{}
	_ resource.ResourceWithImportState = &repositoryResource{}
)

// RepositoryResource is a function that returns a new repository resource
func RepositoryResource() resource.Resource {
	return &repositoryResource{}
}

// repositoryResource is the resource implementation for repositories
type repositoryResource struct {
	client *dbt_cloud.Client
}

// Metadata returns the resource type name
func (r *repositoryResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

// Schema defines the schema for the resource
func (r *repositoryResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = ResourceSchema()
}

// Create creates the resource and sets the initial Terraform state
func (r *repositoryResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Read the plan data
	var plan RepositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract values from plan
	projectID := int(plan.ProjectID.ValueInt64())
	remoteURL := plan.RemoteURL.ValueString()
	isActive := plan.IsActive.ValueBool()
	gitCloneStrategy := plan.GitCloneStrategy.ValueString()
	pullRequestURLTemplate := plan.PullRequestURLTemplate.ValueString()

	// Initialize parameters
	var gitlabProjectID int
	var githubInstallationID int
	var azureProjectID string
	var azureRepositoryID string
	var azureBypassWebhookRegistrationFailure bool

	// Set optional parameters if provided
	if !plan.GitlabProjectID.IsNull() {
		gitlabProjectID = int(plan.GitlabProjectID.ValueInt64())
	}

	if !plan.GithubInstallationID.IsNull() {
		githubInstallationID = int(plan.GithubInstallationID.ValueInt64())
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

	// Map response to model
	plan.ID = types.StringValue(fmt.Sprintf("%d%s%d", repository.ProjectID, dbt_cloud.ID_DELIMITER, *repository.ID))
	plan.RepositoryID = types.Int64Value(int64(*repository.ID))
	plan.IsActive = types.BoolValue(repository.State == dbt_cloud.STATE_ACTIVE)
	plan.ProjectID = types.Int64Value(int64(repository.ProjectID))
	plan.RemoteURL = types.StringValue(repository.RemoteUrl)
	plan.GitCloneStrategy = types.StringValue(repository.GitCloneStrategy)

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

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *repositoryResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Read the current state
	var state RepositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID
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

	// Get the repository
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

	// Update state with values from API response
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

	/// AAD Project ID and Repo ID shouldn't have their state values updated unless it's intentional
	if repository.AzureActiveDirectoryProjectID != nil {
		state.AzureActiveDirectoryProjectID = types.StringValue(*repository.AzureActiveDirectoryProjectID)
	}

	if repository.AzureActiveDirectoryRepositoryID != nil {
		state.AzureActiveDirectoryRepositoryID = types.StringValue(*repository.AzureActiveDirectoryRepositoryID)
	}

	if repository.AzureBypassWebhookRegistrationFailure != nil {
		state.AzureBypassWebhookRegistrationFailure = types.BoolValue(*repository.AzureBypassWebhookRegistrationFailure)
	} else {
		state.AzureBypassWebhookRegistrationFailure = types.BoolValue(false)
	}

	// Set the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success
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

	if hasIsActiveChange || hasPullRequestURLTemplateChange {
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

		_, err = r.client.UpdateRepository(repositoryID, projectID, *repository)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating repository",
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *repositoryResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Read the current state
	var state RepositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID
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

	// Delete the repository
	_, err := r.client.DeleteRepository(repositoryID, projectID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting repository",
			err.Error(),
		)
		return
	}
}

// ImportState imports a resource by ID
func (r *repositoryResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Split the ID into project ID and repository ID
	parts := strings.Split(req.ID, dbt_cloud.ID_DELIMITER)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format project_id%srepository_id, got: %s", dbt_cloud.ID_DELIMITER, req.ID),
		)
		return
	}

	// Set the ID attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// Configure adds the provider configured client to the resource
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
