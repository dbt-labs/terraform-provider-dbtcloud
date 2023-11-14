package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/data_sources"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/resources"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("DBT_CLOUD_TOKEN", nil),
				Description: "API token for your dbt Cloud. Instead of setting the parameter, you can set the environment variable `DBT_CLOUD_TOKEN`",
			},
			"account_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("DBT_CLOUD_ACCOUNT_ID", nil),
				Description: "Account identifier for your dbt Cloud implementation. Instead of setting the parameter, you can set the environment variable `DBT_CLOUD_ACCOUNT_ID`",
			},
			"host_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc(
					"DBT_CLOUD_HOST_URL",
					"https://cloud.getdbt.com/api",
				),
				Description: "URL for your dbt Cloud deployment. Instead of setting the parameter, you can set the environment variable `DBT_CLOUD_HOST_URL` - Defaults to https://cloud.getdbt.com/api",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"dbtcloud_group":                    data_sources.DatasourceGroup(),
			"dbtcloud_job":                      data_sources.DatasourceJob(),
			"dbtcloud_project":                  data_sources.DatasourceProject(),
			"dbtcloud_environment":              data_sources.DatasourceEnvironment(),
			"dbtcloud_environment_variable":     data_sources.DatasourceEnvironmentVariable(),
			"dbtcloud_snowflake_credential":     data_sources.DatasourceSnowflakeCredential(),
			"dbtcloud_bigquery_credential":      data_sources.DatasourceBigQueryCredential(),
			"dbtcloud_postgres_credential":      data_sources.DatasourcePostgresCredential(),
			"dbtcloud_databricks_credential":    data_sources.DatasourceDatabricksCredential(),
			"dbtcloud_connection":               data_sources.DatasourceConnection(),
			"dbtcloud_bigquery_connection":      data_sources.DatasourceBigQueryConnection(),
			"dbtcloud_repository":               data_sources.DatasourceRepository(),
			"dbtcloud_user":                     data_sources.DatasourceUser(),
			"dbtcloud_service_token":            data_sources.DatasourceServiceToken(),
			"dbtcloud_webhook":                  data_sources.DatasourceWebhook(),
			"dbtcloud_privatelink_endpoint":     data_sources.DatasourcePrivatelinkEndpoint(),
			"dbtcloud_notification":             data_sources.DatasourceNotification(),
			"dbtcloud_user_groups":              data_sources.DatasourceUserGroups(),
			"dbtcloud_extended_attributes":      data_sources.DatasourceExtendedAttributes(),
			"dbtcloud_group_users":              data_sources.DatasourceGroupUsers(),
			"dbtcloud_azure_dev_ops_project":    data_sources.DatasourceAzureDevOpsProject(),
			"dbtcloud_azure_dev_ops_repository": data_sources.DatasourceAzureDevOpsRepository(),
			// legacy data sources to remove from 0.3
			"dbt_cloud_group":                 data_sources.DatasourceGroup(),
			"dbt_cloud_job":                   data_sources.DatasourceJob(),
			"dbt_cloud_project":               data_sources.DatasourceProject(),
			"dbt_cloud_environment":           data_sources.DatasourceEnvironment(),
			"dbt_cloud_environment_variable":  data_sources.DatasourceEnvironmentVariable(),
			"dbt_cloud_snowflake_credential":  data_sources.DatasourceSnowflakeCredential(),
			"dbt_cloud_bigquery_credential":   data_sources.DatasourceBigQueryCredential(),
			"dbt_cloud_postgres_credential":   data_sources.DatasourcePostgresCredential(),
			"dbt_cloud_databricks_credential": data_sources.DatasourceDatabricksCredential(),
			"dbt_cloud_connection":            data_sources.DatasourceConnection(),
			"dbt_cloud_bigquery_connection":   data_sources.DatasourceBigQueryConnection(),
			"dbt_cloud_repository":            data_sources.DatasourceRepository(),
			"dbt_cloud_user":                  data_sources.DatasourceUser(),
			"dbt_cloud_service_token":         data_sources.DatasourceServiceToken(),
			"dbt_cloud_webhook":               data_sources.DatasourceWebhook(),
			"dbt_cloud_privatelink_endpoint":  data_sources.DatasourcePrivatelinkEndpoint(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"dbtcloud_job":                   resources.ResourceJob(),
			"dbtcloud_project":               resources.ResourceProject(),
			"dbtcloud_project_connection":    resources.ResourceProjectConnection(),
			"dbtcloud_project_repository":    resources.ResourceProjectRepository(),
			"dbtcloud_project_artefacts":     resources.ResourceProjectArtefacts(),
			"dbtcloud_environment":           resources.ResourceEnvironment(),
			"dbtcloud_environment_variable":  resources.ResourceEnvironmentVariable(),
			"dbtcloud_databricks_credential": resources.ResourceDatabricksCredential(),
			"dbtcloud_snowflake_credential":  resources.ResourceSnowflakeCredential(),
			"dbtcloud_bigquery_credential":   resources.ResourceBigQueryCredential(),
			"dbtcloud_postgres_credential":   resources.ResourcePostgresCredential(),
			"dbtcloud_connection":            resources.ResourceConnection(),
			"dbtcloud_bigquery_connection":   resources.ResourceBigQueryConnection(),
			"dbtcloud_repository":            resources.ResourceRepository(),
			"dbtcloud_group":                 resources.ResourceGroup(),
			"dbtcloud_service_token":         resources.ResourceServiceToken(),
			"dbtcloud_webhook":               resources.ResourceWebhook(),
			"dbtcloud_notification":          resources.ResourceNotification(),
			"dbtcloud_user_groups":           resources.ResourceUserGroups(),
			"dbtcloud_license_map":           resources.ResourceLicenseMap(),
			"dbtcloud_extended_attributes":   resources.ResourceExtendedAttributes(),
			// legacy resources to remove from 0.3
			"dbt_cloud_job":                   resources.ResourceJob(),
			"dbt_cloud_project":               resources.ResourceProject(),
			"dbt_cloud_project_connection":    resources.ResourceProjectConnection(),
			"dbt_cloud_project_repository":    resources.ResourceProjectRepository(),
			"dbt_cloud_project_artefacts":     resources.ResourceProjectArtefacts(),
			"dbt_cloud_environment":           resources.ResourceEnvironment(),
			"dbt_cloud_environment_variable":  resources.ResourceEnvironmentVariable(),
			"dbt_cloud_databricks_credential": resources.ResourceDatabricksCredential(),
			"dbt_cloud_snowflake_credential":  resources.ResourceSnowflakeCredential(),
			"dbt_cloud_bigquery_credential":   resources.ResourceBigQueryCredential(),
			"dbt_cloud_postgres_credential":   resources.ResourcePostgresCredential(),
			"dbt_cloud_connection":            resources.ResourceConnection(),
			"dbt_cloud_bigquery_connection":   resources.ResourceBigQueryConnection(),
			"dbt_cloud_repository":            resources.ResourceRepository(),
			"dbt_cloud_group":                 resources.ResourceGroup(),
			"dbt_cloud_service_token":         resources.ResourceServiceToken(),
			"dbt_cloud_webhook":               resources.ResourceWebhook(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(
	ctx context.Context,
	d *schema.ResourceData,
) (interface{}, diag.Diagnostics) {

	token := d.Get("token").(string)
	account_id := d.Get("account_id").(int)
	host_url := d.Get("host_url").(string)

	var diags diag.Diagnostics

	if (token != "") && (account_id != 0) {
		c, err := dbt_cloud.NewClient(&account_id, &token, &host_url)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to login to dbt Cloud",
				Detail:   err.Error(),
			})
			return nil, diags
		}

		return c, diags
	}

	c, err := dbt_cloud.NewClient(nil, nil, nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create dbt Cloud client",
			Detail:   err.Error(),
		})
		return nil, diags
	}

	return c, diags
}
