package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/sdkv2/data_sources"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/sdkv2/resources"
)

func SDKProvider(version string) func() *schema.Provider {
	return func() *schema.Provider {

		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"token": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "API token for your dbt Cloud. Instead of setting the parameter, you can set the environment variable `DBT_CLOUD_TOKEN`",
				},
				"account_id": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "Account identifier for your dbt Cloud implementation. Instead of setting the parameter, you can set the environment variable `DBT_CLOUD_ACCOUNT_ID`",
				},
				"host_url": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "URL for your dbt Cloud deployment. Instead of setting the parameter, you can set the environment variable `DBT_CLOUD_HOST_URL` - Defaults to https://cloud.getdbt.com/api",
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				// "dbtcloud_job":                   data_sources.DatasourceJob(),
				"dbtcloud_project":               data_sources.DatasourceProject(),
				"dbtcloud_environment_variable":  data_sources.DatasourceEnvironmentVariable(),
				"dbtcloud_snowflake_credential":  data_sources.DatasourceSnowflakeCredential(),
				"dbtcloud_bigquery_credential":   data_sources.DatasourceBigQueryCredential(),
				"dbtcloud_postgres_credential":   data_sources.DatasourcePostgresCredential(),
				"dbtcloud_databricks_credential": data_sources.DatasourceDatabricksCredential(),
				"dbtcloud_connection":            data_sources.DatasourceConnection(),
				"dbtcloud_bigquery_connection":   data_sources.DatasourceBigQueryConnection(),
				"dbtcloud_repository":            data_sources.DatasourceRepository(),
				"dbtcloud_webhook":               data_sources.DatasourceWebhook(),
				"dbtcloud_privatelink_endpoint":  data_sources.DatasourcePrivatelinkEndpoint(),
				"dbtcloud_user_groups":           data_sources.DatasourceUserGroups(),
				"dbtcloud_extended_attributes":   data_sources.DatasourceExtendedAttributes(),
				"dbtcloud_group_users":           data_sources.DatasourceGroupUsers(),
			},
			ResourcesMap: map[string]*schema.Resource{
				// "dbtcloud_job":                               resources.ResourceJob(),
				"dbtcloud_project":                           resources.ResourceProject(),
				"dbtcloud_project_connection":                resources.ResourceProjectConnection(),
				"dbtcloud_project_repository":                resources.ResourceProjectRepository(),
				"dbtcloud_project_artefacts":                 resources.ResourceProjectArtefacts(),
				"dbtcloud_environment":                       resources.ResourceEnvironment(),
				"dbtcloud_environment_variable":              resources.ResourceEnvironmentVariable(),
				"dbtcloud_databricks_credential":             resources.ResourceDatabricksCredential(),
				"dbtcloud_snowflake_credential":              resources.ResourceSnowflakeCredential(),
				"dbtcloud_bigquery_credential":               resources.ResourceBigQueryCredential(),
				"dbtcloud_postgres_credential":               resources.ResourcePostgresCredential(),
				"dbtcloud_connection":                        resources.ResourceConnection(),
				"dbtcloud_bigquery_connection":               resources.ResourceBigQueryConnection(),
				"dbtcloud_repository":                        resources.ResourceRepository(),
				"dbtcloud_webhook":                           resources.ResourceWebhook(),
				"dbtcloud_user_groups":                       resources.ResourceUserGroups(),
				"dbtcloud_extended_attributes":               resources.ResourceExtendedAttributes(),
				"dbtcloud_environment_variable_job_override": resources.ResourceEnvironmentVariableJobOverride(),
				"dbtcloud_fabric_connection":                 resources.ResourceFabricConnection(),
				"dbtcloud_fabric_credential":                 resources.ResourceFabricCredential(),
			},
			ConfigureContextFunc: providerConfigure,
		}
		return p
	}
}

func providerConfigure(
	ctx context.Context,
	d *schema.ResourceData,
) (interface{}, diag.Diagnostics) {

	account_id := d.Get("account_id").(int)
	token := d.Get("token").(string)
	host_url := d.Get("host_url").(string)

	if account_id == 0 {
		accountIDString := os.Getenv("DBT_CLOUD_ACCOUNT_ID")
		account_id, _ = strconv.Atoi(accountIDString)
	}

	if token == "" {
		token = os.Getenv("DBT_CLOUD_TOKEN")
	}

	if host_url == "" {
		host_url = os.Getenv("DBT_CLOUD_HOST_URL")
		if host_url == "" {
			host_url = "https://cloud.getdbt.com/api"
		}
	}

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
