package data_sources

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var snowflakeCredentialSchema = map[string]*schema.Schema{
	"project_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID",
	},
	"credential_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Credential ID",
	},
	"is_active": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the Snowflake credential is active",
	},
	"auth_type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The type of Snowflake credential ('password' or 'keypair')",
	},
	"schema": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Default schema name",
	},
	"user": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Username for Snowflake",
	},
	"num_threads": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Number of threads to use",
	},
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
	if err := d.Set("num_threads", snowflakeCredential.Threads); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", snowflakeCredential.Project_Id, dbt_cloud.ID_DELIMITER, *snowflakeCredential.ID))

	return diags
}
