package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceBigQueryCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigQueryCredentialCreate,
		ReadContext:   resourceBigQueryCredentialRead,
		UpdateContext: resourceBigQueryCredentialUpdate,
		DeleteContext: resourceBigQueryCredentialDelete,

		Schema: map[string]*schema.Schema{
			"is_active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the BigQuery credential is active",
			},
			"project_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID to create the BigQuery credential in",
			},
			"credential_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The system BigQuery credential ID",
			},
			"dataset": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Default dataset name",
			},
			"num_threads": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Number of threads to use",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceBigQueryCredentialCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	isActive := d.Get("is_active").(bool)
	projectId := d.Get("project_id").(int)
	dataset := d.Get("dataset").(string)
	numThreads := d.Get("num_threads").(int)

	BigQueryCredential, err := c.CreateBigQueryCredential(
		projectId,
		"bigquery",
		isActive,
		dataset,
		numThreads,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(
		fmt.Sprintf(
			"%d%s%d",
			BigQueryCredential.Project_Id,
			dbt_cloud.ID_DELIMITER,
			*BigQueryCredential.ID,
		),
	)

	resourceBigQueryCredentialRead(ctx, d, m)

	return diags
}

func resourceBigQueryCredentialRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectId, BigQueryCredentialId, err := helper.SplitIDToInts(
		d.Id(),
		"dbtcloud_bigquery_credential",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	BigQueryCredential, err := c.GetBigQueryCredential(projectId, BigQueryCredentialId)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("credential_id", BigQueryCredentialId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_active", BigQueryCredential.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", BigQueryCredential.Project_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dataset", BigQueryCredential.Dataset); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("num_threads", BigQueryCredential.Threads); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceBigQueryCredentialUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectId, BigQueryCredentialId, err := helper.SplitIDToInts(
		d.Id(),
		"dbtcloud_bigquery_credential",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("dataset") || d.HasChange("num_threads") {
		BigQueryCredential, err := c.GetBigQueryCredential(projectId, BigQueryCredentialId)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("dataset") {
			dataset := d.Get("dataset").(string)
			BigQueryCredential.Dataset = dataset
		}
		if d.HasChange("num_threads") {
			numThreads := d.Get("num_threads").(int)
			BigQueryCredential.Threads = numThreads
		}

		_, err = c.UpdateBigQueryCredential(projectId, BigQueryCredentialId, *BigQueryCredential)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceBigQueryCredentialRead(ctx, d, m)
}

func resourceBigQueryCredentialDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString, BigQueryCredentialIdString, err := helper.SplitIDToStrings(
		d.Id(),
		"dbtcloud_bigquery_credential",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.DeleteCredential(BigQueryCredentialIdString, projectIdString)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
