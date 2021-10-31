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
		},
		DataSourcesMap: map[string]*schema.Resource{
			"dbt_cloud_job":                  data_sources.DatasourceJob(),
			"dbt_cloud_project":              data_sources.DatasourceProject(),
			"dbt_cloud_environment":          data_sources.DatasourceEnvironment(),
			"dbt_cloud_snowflake_credential": data_sources.DatasourceSnowflakeCredential(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"dbt_cloud_job":                  resources.ResourceJob(),
			"dbt_cloud_project":              resources.ResourceProject(),
			"dbt_cloud_environment":          resources.ResourceEnvironment(),
			"dbt_cloud_snowflake_credential": resources.ResourceSnowflakeCredential(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	token := d.Get("token").(string)
	account_id := d.Get("account_id").(int)

	var diags diag.Diagnostics

	if (token != "") && (account_id != 0) {
		c, err := dbt_cloud.NewClient(&account_id, &token)

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

	c, err := dbt_cloud.NewClient(nil, nil)
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
