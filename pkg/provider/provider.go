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
			"dbt_cloud_job":                  data_sources.DatasourceJob(),
			"dbt_cloud_project":              data_sources.DatasourceProject(),
			"dbt_cloud_environment":          data_sources.DatasourceEnvironment(),
			"dbt_cloud_snowflake_credential": data_sources.DatasourceSnowflakeCredential(),
			"dbt_cloud_connection":           data_sources.DatasourceConnection(),
			"dbt_cloud_repository":           data_sources.DatasourceRepository(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"dbt_cloud_job":                  resources.ResourceJob(),
			"dbt_cloud_project":              resources.ResourceProject(),
			"dbt_cloud_environment":          resources.ResourceEnvironment(),
			"dbt_cloud_environment_variable": resources.ResourceEnvironmentVariable(),
			"dbt_cloud_snowflake_credential": resources.ResourceSnowflakeCredential(),
			"dbt_cloud_connection":           resources.ResourceConnection(),
			"dbt_cloud_repository":           resources.ResourceRepository(),
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
