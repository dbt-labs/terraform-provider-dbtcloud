package data_sources

import (
	"context"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var serviceTokenSchema = map[string]*schema.Schema{
	"service_token_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "ID of the service token",
	},
	"name": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Service token name",
	},
	"uid": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The UID of the service token (part of the token secret)",
	},
	"service_token_permissions": &schema.Schema{
		Type:        schema.TypeSet,
		Computed:    true,
		Description: "Permissions set for the service token",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"permission_set": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Set of permissions to apply",
				},
				"project_id": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Project ID to apply this permission to for this service token",
				},
				"all_projects": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Whether or not to apply this permission to all projects for this service token",
				},
			},
		},
	}}

func DatasourceServiceToken() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceTokenRead,
		Schema:      serviceTokenSchema,
	}
}

func datasourceServiceTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	serviceTokenID := d.Get("service_token_id").(int)

	serviceToken, err := c.GetServiceToken(serviceTokenID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", serviceToken.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("uid", serviceToken.UID); err != nil {
		return diag.FromErr(err)
	}

	permissions := make([]interface{}, len(serviceToken.Permissions))
	for i, permission := range serviceToken.Permissions {
		p := make(map[string]interface{})
		p["permission_set"] = permission.Set
		p["project_id"] = permission.ProjectID
		p["all_projects"] = permission.AllProjects
		permissions[i] = p
	}
	if err := d.Set("service_token_permissions", permissions); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(*serviceToken.ID))

	return diags
}
