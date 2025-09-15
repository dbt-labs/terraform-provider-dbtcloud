package provider

import (
	"context"
	"os"
	"regexp"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/environment_variable"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/environment_variable_job_override"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/extended_attributes"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/group_users"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/partial_environment_variable"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/privatelink_endpoint"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/runs"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/semantic_layer_configuration"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/semantic_layer_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/snowflake_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/synapse_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/teradata_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/user_groups"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/account_features"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/athena_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/azure_dev_ops_project"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/azure_dev_ops_repository"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/bigquery_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/databricks_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/environment"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/fabric_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/global_connection"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/group"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/group_partial_permissions"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/ip_restrictions_rule"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/job"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/license_map"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/lineage_integration"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/model_notifications"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/notification"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/oauth_configuration"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/partial_license_map"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/partial_notification"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/postgres_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/project"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/project_artefacts"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/project_repository"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/redshift_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/repository"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/scim_group_permissions"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/semantic_layer_credential_service_token_mapping"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/service_token"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/starburst_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/user"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/webhook"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ provider.Provider = &dbtCloudProvider{}
)

func New() provider.Provider {
	return &dbtCloudProvider{}
}

type dbtCloudProvider struct{}

func (p *dbtCloudProvider) Metadata(
	_ context.Context,
	_ provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "dbtcloud"
}

func (p *dbtCloudProvider) Schema(
	_ context.Context,
	_ provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "API token for your dbt Cloud. Instead of setting the parameter, you can set the environment variable `DBT_CLOUD_TOKEN`",
			},
			"account_id": schema.Int64Attribute{
				Optional:    true,
				Description: "Account identifier for your dbt Cloud implementation. Instead of setting the parameter, you can set the environment variable `DBT_CLOUD_ACCOUNT_ID`",
			},
			"host_url": schema.StringAttribute{
				Optional:    true,
				Description: "URL for your dbt Cloud deployment. Instead of setting the parameter, you can set the environment variable `DBT_CLOUD_HOST_URL` - Defaults to https://cloud.getdbt.com/api",
			},
			"retry_interval_seconds": schema.Int64Attribute{
				Optional:    true,
				Description: "The number of seconds to wait before retrying a request that failed due to rate limiting. Defaults to 10 seconds.",
			},
			"max_retries": schema.Int64Attribute{
				Optional:    true,
				Description: "The maximum number of retries to attempt for requests that fail due to rate limiting. Defaults to 3 retries.",
			},
			"disable_retry": schema.BoolAttribute{
				Optional:    true,
				Description: "If set to true, the provider will not retry requests that fail due to rate limiting. Defaults to false.",
			},
			"retriable_status_codes": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of HTTP status codes that should be retried when encountered. Defaults to [429, 500, 502, 503, 504].",
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^[1-5][0-9][0-9]$`),
							"must be a valid HTTP status code (100-599)",
						),
					),
				},
			},
		},
	}
}

type dbtCloudProviderModel struct {
	Token                types.String `tfsdk:"token"`
	AccountID            types.Int64  `tfsdk:"account_id"`
	HostURL              types.String `tfsdk:"host_url"`
	MaxRetries           types.Int64  `tfsdk:"max_retries"`
	RetryIntervalSeconds types.Int64  `tfsdk:"retry_interval_seconds"`
	DisableRetry         types.Bool   `tfsdk:"disable_retry"`
	RetriableStatusCodes types.List   `tfsdk:"retriable_status_codes"`
}

func (p *dbtCloudProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	tflog.Info(ctx, "Configuring dbt Cloud client")

	var config dbtCloudProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.AccountID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("account_id"),
			"Unknown dbt Cloud account identifier",
			"dbt Cloud account identifier must be provided in order to establish a connection",
		)
	}

	if config.HostURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host_url"),
			"Unknown dbt Cloud host URL",
			"dbt Cloud host URL must be provided in order to establish a connection",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown dbt Cloud token",
			"Token must be provided in order to establish a connection",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	accountIDString := os.Getenv("DBT_CLOUD_ACCOUNT_ID")
	accountID, _ := strconv.Atoi(accountIDString)
	token := os.Getenv("DBT_CLOUD_TOKEN")
	hostURL := os.Getenv("DBT_CLOUD_HOST_URL")

	if !config.AccountID.IsNull() {
		accountID = int(config.AccountID.ValueInt64())
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	if !config.HostURL.IsNull() {
		hostURL = config.HostURL.ValueString()
	}

	if accountID == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("account_id"),
			"Missing dbt Cloud account identifier",
			"dbt Cloud account identifier must be provided in order to establish a connection",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing dbt Cloud token",
			"dbt Cloud token must be provided in order to establish a connection, currently no default is set",
		)
	}

	if hostURL == "" {
		hostURL = "https://cloud.getdbt.com/api"
	}

	if resp.Diagnostics.HasError() {
		return
	}
	maxRetries := 3
	if !config.MaxRetries.IsNull() {
		maxRetries = int(config.MaxRetries.ValueInt64())
	}

	retryIntervalSeconds := 10
	if !config.RetryIntervalSeconds.IsNull() {
		retryIntervalSeconds = int(config.RetryIntervalSeconds.ValueInt64())
	}

	if config.DisableRetry.ValueBool() {
		maxRetries = 1
		retryIntervalSeconds = 0
	}

	retriableStatusCodes := []string{"429", "500", "502", "503", "504"}
	if !config.RetriableStatusCodes.IsNull() {
		retriableStatusCodes = make([]string, len(config.RetriableStatusCodes.Elements()))
		for i, elem := range config.RetriableStatusCodes.Elements() {
			strElem, ok := elem.(types.String)
			if !ok {
				resp.Diagnostics.AddError(
					"Invalid Retriable Status Codes",
					"All elements in the retriable_status_codes list must be strings. "+
						"Element at index "+strconv.Itoa(i)+" is of an invalid type.",
				)
				return
			}
			retriableStatusCodes[i] = strElem.ValueString()
		}
	}

	client, err := dbt_cloud.NewClient(&accountID, &token, &hostURL, &maxRetries, &retryIntervalSeconds, retriableStatusCodes)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create dbt Cloud API Client",
			"An unexpected error occurred when creating the dbt Cloud API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"dbt Cloud API Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured dbt Cloud client", map[string]any{"success": true})
}

func (p *dbtCloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		athena_credential.NewAthenaCredentialDataSource,
		azure_dev_ops_project.AzureDevOpsProjectDataSource,
		azure_dev_ops_repository.AzureDevOpsRepositoryDataSource,
		environment.EnvironmentDataSource,
		environment.EnvironmentsDataSource,
		global_connection.GlobalConnectionDataSource,
		global_connection.GlobalConnectionsDataSource,
		group.GroupDataSource,
		group.GroupsDataSource,
		job.JobDataSource,
		job.JobsDataSource,
		model_notifications.ModelNotificationsDataSource,
		notification.NotificationDataSource,
		project.ProjectsDataSource,
		repository.RepositoryDataSource,
		service_token.ServiceTokenDataSource,
		starburst_credential.StarburstCredentialDataSource,
		user.UserDataSource,
		user.UsersDataSource,
		bigquery_credential.BigqueryCredentialDataSource,
		redshift_credential.RedshiftCredentialDataSource,
		postgres_credential.PostgresCredentialDataSource,
		user_groups.UserGroupDataSource,
		webhook.WebhookDataSource,
		databricks_credential.DatabricksCredentialDataSource,
		snowflake_credential.SnowflakeCredentialDataSource,
		extended_attributes.ExtendedAttributesDataSource,
		teradata_credential.TeradataCredentialDataSource,
		environment_variable.EnvironmentVariableDataSource,
		project.ProjectDataSource,
		privatelink_endpoint.PrivatelinkEndpointDataSource,
		privatelink_endpoint.PrivatelinkEndpointDataSourceAll,
		group_users.GroupUsersDataSource,
		runs.RunsDataSource,
		synapse_credential.SynapseCredentialDataSource,
	}
}

func (p *dbtCloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		account_features.AccountFeaturesResource,
		athena_credential.NewAthenaCredentialResource,
		global_connection.GlobalConnectionResource,
		group_partial_permissions.GroupPartialPermissionsResource,
		group.GroupResource,
		ip_restrictions_rule.IPRestrictionsRuleResource,
		license_map.LicenseMapResource,
		lineage_integration.LineageIntegrationResource,
		model_notifications.ModelNotificationsResource,
		notification.NotificationResource,
		oauth_configuration.OAuthConfigurationResource,
		partial_environment_variable.PartialEnvironmentVariableResource,
		partial_license_map.PartialLicenseMapResource,
		partial_notification.PartialNotificationResource,
		project_artefacts.ProjectArtefactsResource,
		repository.RepositoryResource,
		scim_group_permissions.ScimGroupPermissionsResource,
		service_token.ServiceTokenResource,
		starburst_credential.StarburstCredentialResource,
		bigquery_credential.BigqueryCredentialResource,
		redshift_credential.RedshiftCredentialResource,
		postgres_credential.PostgresCredentialResource,
		fabric_credential.FabricCredentialResource,
		user_groups.UserGroupsResource,
		webhook.WebhookResource,
		databricks_credential.DatabricksCredentialResource,
		environment.EnvironmentResource,
		snowflake_credential.SnowflakeCredentialResource,
		extended_attributes.ExtendedAttributesResource,
		teradata_credential.TeradataCredentialResource,
		job.JobResource,
		project_repository.ProjectRepositoryResource,
		environment_variable.EnvironmentVariableResource,
		environment_variable_job_override.EnvironmentVariableJobOverrideResource,
		project.ProjectResource,
		semantic_layer_configuration.SemanticLayerConfigurationResource,
		semantic_layer_credential_service_token_mapping.SemanticLayerCredentialServiceTokenMappingResource,
		semantic_layer_credential.SnowflakeSemanticLayerCredentialResource,
		semantic_layer_credential.BigQuerySemanticLayerCredentialResource,
		semantic_layer_credential.RedshiftSemanticLayerCredentialResource,
		semantic_layer_credential.DatabricksSemanticLayerCredentialResource,
		semantic_layer_credential.PostgresSemanticLayerCredentialResource,
		synapse_credential.SynapseCredentialResource,
	}
}
