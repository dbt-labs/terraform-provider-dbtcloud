package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/data_sources"
	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/resources"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("DBT_CLOUD_TOKEN", nil),
			},
			"account_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("DBT_CLOUD_ACCOUNT_ID", nil),
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"dbt_cloud_job": data_sources.DatasourceJob(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"dbt_cloud_job": resources.ResourceJob(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	token := d.Get("token").(string)

	var diags diag.Diagnostics

	if token != "" {
		url := fmt.Sprintf("%s/accounts/", "https://cloud.getdbt.com/api/v2/")
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))
		resp, err := client.Do(req)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to login to DBT Cloud",
				Detail:   "Unable to auth token for DBT Cloud",
			})
			return nil, diags
		}

		return resp, diags
	}

	return nil, nil
}
