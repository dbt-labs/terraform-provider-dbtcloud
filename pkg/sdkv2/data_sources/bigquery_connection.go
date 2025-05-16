package data_sources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var bigQueryConnectionSchema = map[string]*schema.Schema{
	"connection_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Connection Identifier",
	},
	"project_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID to create the connection in",
	},
	"is_active": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the connection is active",
	},
	"name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Connection name",
	},
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The type of connection",
	},
	// field in details
	"gcp_project_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "GCP project ID",
	},
	"timeout_seconds": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Timeout in seconds for queries",
	},
	"private_key_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Private key ID of the Service Account",
	},
	"private_key": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Private key of the Service Account",
	},
	"client_email": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Service Account email",
	},
	"client_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Client ID of the Service Account",
	},
	"auth_uri": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Auth URI for the Service Account",
	},
	"token_uri": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Token URI for the Service Account",
	},
	"auth_provider_x509_cert_url": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Auth Provider X509 Cert URL for the Service Account",
	},
	"client_x509_cert_url": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Client X509 Cert URL for the Service Account",
	},
	"retries": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Number of retries for queries",
	},
	"location": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Location to create new Datasets in",
	},
	"maximum_bytes_billed": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Max number of bytes that can be billed for a given BigQuery query",
	},
	"execution_project": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Project to bill for query execution",
	},
	"priority": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The priority with which to execute BigQuery queries",
	},
	"gcs_bucket": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "URI for a Google Cloud Storage bucket to host Python code executed via Datapro",
	},
	"dataproc_region": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Google Cloud region for PySpark workloads on Dataproc",
	},
	"dataproc_cluster_name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Dataproc cluster name for PySpark workloads",
	},
	"is_configured_for_oauth": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the connection is configured for OAuth or not",
	},
}

func DatasourceBigQueryConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceBigQueryConnectionRead,
		Schema:      bigQueryConnectionSchema,
	}
}

func datasourceBigQueryConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	connectionID := d.Get("connection_id").(int)
	projectID := d.Get("project_id").(int)

	connection, err := c.GetBigQueryConnection(strconv.Itoa(connectionID), strconv.Itoa(projectID))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("is_active", connection.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", connection.ProjectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", connection.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", connection.Type); err != nil {
		return diag.FromErr(err)
	}
	// under details
	if err := d.Set("gcp_project_id", connection.Details.GcpProjectId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("timeout_seconds", connection.Details.TimeoutSeconds); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("private_key_id", connection.Details.PrivateKeyId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("client_email", connection.Details.ClientEmail); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("client_id", connection.Details.ClientId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("auth_uri", connection.Details.AuthUri); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("token_uri", connection.Details.TokenUri); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("auth_provider_x509_cert_url", connection.Details.AuthProviderX509CertUrl); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("client_x509_cert_url", connection.Details.ClientX509CertUrl); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("retries", connection.Details.Retries); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("location", connection.Details.Location); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("maximum_bytes_billed", connection.Details.MaximumBytesBilled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("execution_project", connection.Details.ExecutionProject); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("priority", connection.Details.Priority); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("gcs_bucket", connection.Details.GcsBucket); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dataproc_region", connection.Details.DataprocRegion); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dataproc_cluster_name", connection.Details.DataprocClusterName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_configured_for_oauth", connection.Details.IsConfiguredOAuth); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", connection.ProjectID, dbt_cloud.ID_DELIMITER, *connection.ID))

	return diags
}
