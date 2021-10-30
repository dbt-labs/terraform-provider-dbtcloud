package data_sources

import (
	"context"
	"fmt"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var snowflakeCredentialSchema = map[string]*schema.Schema{
	"is_active": &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "Whether the Snowflake credential is active",
	},
	"project_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID to create the Snowflake credential in",
	},
	"auth_type": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The type of Snowflake credential ('password' only currently supported in Terraform)",
		ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
			type_ := val.(string)
			switch type_ {
			case
				"password":
				return
			}
			errs = append(errs, fmt.Errorf("%q must be password, got: %q", key, type_))
			return
		},
	},
	"schema": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "Default schema name",
	},
	"user": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "Username for Snowflake",
	},
	"password": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "Password for Snowflake",
	},
	"num_threads": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Number of threads to use",
	},
	// TODO: add private_key and private_key_passphrase
}

func DatasourceSnowflakeCredential() *schema.Resource {
	return &schema.Resource{
		ReadContext: snowflakeCredentialRead,
		Schema:      snowflakeCredentialSchema,
	}
}

func snowflakeCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	credentialID := d.Get("credential_id").(int)
	projectID := d.Get("project_id").(int)

	snowflakeCredential, err := c.GetSnowflakeCredential(projectID, credentialID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("is_active", snowflakeCredential.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", snowflakeCredential.Project_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("auth_type", snowflakeCredential.Auth_Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("schema", snowflakeCredential.Schema); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("user", snowflakeCredential.User); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("password", snowflakeCredential.Password); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("num_threads", snowflakeCredential.Threads); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", snowflakeCredential.Project_Id, dbt_cloud.ID_DELIMITER, *snowflakeCredential.ID))

	return diags
}
