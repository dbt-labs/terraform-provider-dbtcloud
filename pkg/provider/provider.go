package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/data_sources"
	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/resources"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("DBT_CLOUD_TOKEN", nil),
				Description: "API token for your DBT Cloud",
			},
			"account_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("DBT_CLOUD_ACCOUNT_ID", nil),
				Description: "Account identifier for your DBT Cloud implementation",
			},
			"host_url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DBT_CLOUD_HOST_URL", "https://cloud.getdbt.com/api"),
				Description: "URL for your DBT Cloud deployment",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"dbt-cloud_group":                data_sources.DatasourceGroup(),
			"dbt-cloud_job":                  data_sources.DatasourceJob(),
			"dbt-cloud_project":              data_sources.DatasourceProject(),
			"dbt-cloud_environment":          data_sources.DatasourceEnvironment(),
			"dbt-cloud_environment_variable": data_sources.DatasourceEnvironmentVariable(),
			"dbt-cloud_snowflake_credential": data_sources.DatasourceSnowflakeCredential(),
			"dbt-cloud_connection":           data_sources.DatasourceConnection(),
			"dbt-cloud_repository":           data_sources.DatasourceRepository(),
			"dbt-cloud_user":                 data_sources.DatasourceUser(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"dbt-cloud_job":                   resources.ResourceJob(),
			"dbt-cloud_project":               resources.ResourceProject(),
			"dbt-cloud_project_connection":    resources.ResourceProjectConnection(),
			"dbt-cloud_project_repository":    resources.ResourceProjectRepository(),
			"dbt-cloud_environment":           resources.ResourceEnvironment(),
			"dbt-cloud_environment_variable":  resources.ResourceEnvironmentVariable(),
			"dbt-cloud_databricks_credential": resources.ResourceDatabricksCredential(),
			"dbt-cloud_snowflake_credential":  resources.ResourceSnowflakeCredential(),
			"dbt-cloud_connection":            resources.ResourceConnection(),
			"dbt-cloud_repository":            resources.ResourceRepository(),
			"dbt-cloud_group":                 resources.ResourceGroup(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	token := d.Get("token").(string)
	account_id := d.Get("account_id").(int)
	host_url := d.Get("host_url").(string)

	var diags diag.Diagnostics

	if (token != "") && (account_id != 0) {
		c, err := dbt_cloud.NewClient(&account_id, &token, &host_url)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to login to DBT Cloud",
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
			Summary:  "Unable to create DBT Cloud client",
			Detail:   err.Error(),
		})
		return nil, diags
	}

	return c, diags
}
