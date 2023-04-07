package data_sources

import (
	"context"
	"fmt"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var postgresCredentialSchema = map[string]*schema.Schema{
	"project_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID",
	},
	"credential_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Credential ID",
	},
	"is_active": &schema.Schema{
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the Postgres credential is active",
	},
	"default_schema": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Default schema name",
	},
	"username": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Username for Postgres",
	},
	"num_threads": &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Number of threads to use",
	},
}

func DatasourcePostgresCredential() *schema.Resource {
	return &schema.Resource{
		ReadContext: postgresCredentialRead,
		Schema:      postgresCredentialSchema,
	}
}

func postgresCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	credentialID := d.Get("credential_id").(int)
	projectID := d.Get("project_id").(int)

	postgresCredential, err := c.GetPostgresCredential(projectID, credentialID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("is_active", postgresCredential.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", postgresCredential.Project_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_schema", postgresCredential.Default_Schema); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("username", postgresCredential.Username); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("num_threads", postgresCredential.Threads); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", postgresCredential.Project_Id, dbt_cloud.ID_DELIMITER, *postgresCredential.ID))

	return diags
}
