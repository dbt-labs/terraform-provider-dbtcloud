package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ datasource.DataSource              = &repositoryDataSource{}
	_ datasource.DataSourceWithConfigure = &repositoryDataSource{}
)

// RepositoryDataSource returns a new repository data source
func RepositoryDataSource() datasource.DataSource {
	return &repositoryDataSource{}
}

// repositoryDataSource is the data source implementation for repositories
type repositoryDataSource struct {
	client *dbt_cloud.Client
}

// Metadata returns the data source type name
func (d *repositoryDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

// Schema defines the schema for the data source
func (d *repositoryDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = DataSourceSchema()
}

// Read fetches the data from the API
func (d *repositoryDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data RepositoryDataSourceModel

	// Read the datasource inputs
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repositoryID := int(data.RepositoryID.ValueInt64())
	projectID := int(data.ProjectID.ValueInt64())

	repository, err := d.client.GetRepository(
		strconv.Itoa(repositoryID),
		strconv.Itoa(projectID),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting repository",
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d%s%d", repository.ProjectID, dbt_cloud.ID_DELIMITER, *repository.ID))
	data.IsActive = types.BoolValue(repository.State == dbt_cloud.STATE_ACTIVE)
	data.ProjectID = types.Int64Value(int64(repository.ProjectID))
	data.RepositoryID = types.Int64Value(int64(*repository.ID))
	data.RemoteURL = types.StringValue(repository.RemoteUrl)
	data.GitCloneStrategy = types.StringValue(repository.GitCloneStrategy)

	if repository.RepositoryCredentialsID != nil {
		data.RepositoryCredentialsID = types.Int64Value(int64(*repository.RepositoryCredentialsID))
	} else {
		data.RepositoryCredentialsID = types.Int64Null()
	}

	if repository.GitlabProjectID != nil {
		data.GitlabProjectID = types.Int64Value(int64(*repository.GitlabProjectID))
	} else {
		data.GitlabProjectID = types.Int64Null()
	}

	if repository.GithubInstallationID != nil {
		data.GithubInstallationID = types.Int64Value(int64(*repository.GithubInstallationID))
	} else {
		data.GithubInstallationID = types.Int64Null()
	}

	if repository.DeployKey != nil {
		data.DeployKey = types.StringValue(repository.DeployKey.PublicKey)
	} else {
		data.DeployKey = types.StringNull()
	}

	if repository.PullRequestURLTemplate != "" {
		data.PullRequestURLTemplate = types.StringValue(repository.PullRequestURLTemplate)
	} else {
		data.PullRequestURLTemplate = types.StringNull()
	}

	if repository.AzureActiveDirectoryProjectID != nil {
		data.AzureActiveDirectoryProjectID = types.StringValue(*repository.AzureActiveDirectoryProjectID)
	} else {
		data.AzureActiveDirectoryProjectID = types.StringNull()
	}

	if repository.AzureActiveDirectoryRepositoryID != nil {
		data.AzureActiveDirectoryRepositoryID = types.StringValue(*repository.AzureActiveDirectoryRepositoryID)
	} else {
		data.AzureActiveDirectoryRepositoryID = types.StringNull()
	}

	if repository.AzureBypassWebhookRegistrationFailure != nil {
		data.AzureBypassWebhookRegistrationFailure = types.BoolValue(*repository.AzureBypassWebhookRegistrationFailure)
	} else {
		data.AzureBypassWebhookRegistrationFailure = types.BoolNull()
	}

	// Set the response data
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source
func (d *repositoryDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}
