package data_sources

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var bigqueryCredentialSchema = map[string]*schema.Schema{
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
		Description: "Whether the BigQuery credential is active",
	},
	"dataset": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Default dataset name",
	},
	"num_threads": &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Number of threads to use",
	},
}

func DatasourceBigQueryCredential() *schema.Resource {
	return &schema.Resource{
		ReadContext: bigqueryCredentialRead,
		Schema:      bigqueryCredentialSchema,
	}
}

func bigqueryCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	credentialID := d.Get("credential_id").(int)
	projectID := d.Get("project_id").(int)

	bigqueryCredential, err := c.GetBigQueryCredential(projectID, credentialID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("is_active", bigqueryCredential.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", bigqueryCredential.Project_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dataset", bigqueryCredential.Dataset); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("num_threads", bigqueryCredential.Threads); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", bigqueryCredential.Project_Id, dbt_cloud.ID_DELIMITER, *bigqueryCredential.ID))

	return diags
}
