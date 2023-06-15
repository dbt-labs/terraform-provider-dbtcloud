package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceBigQueryConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBigQueryConnectionCreate,
		ReadContext:   resourceBigQueryConnectionRead,
		UpdateContext: resourceBigQueryConnectionUpdate,
		DeleteContext: resourceBigQueryConnectionDelete,

		Schema: map[string]*schema.Schema{
			"connection_id": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Connection Identifier",
			},
			"is_active": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the connection is active",
			},
			"project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Project ID to create the connection in",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Connection name",
			},
			// TODO auto-add type
			"type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The type of connection",
				ValidateFunc: validation.StringInSlice(connectionTypes, false),
			},
			// under details
			"gcp_project_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "GCP project ID",
			},
			"timeout_seconds": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Timeout in seconds for queries",
			},
			"private_key_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Private key ID of the Service Account",
			},
			"private_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Private key of the Service Account",
			},
			"client_email": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service Account email",
			},
			"client_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Client ID of the Service Account",
			},
			"auth_uri": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Auth URI for the Service Account",
			},
			"token_uri": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Token URI for the Service Account",
			},
			"auth_provider_x509_cert_url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Auth Provider X509 Cert URL for the Service Account",
			},
			"client_x509_cert_url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Client X509 Cert URL for the Service Account",
			},
			"retries": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Number of retries for queries",
			},
			"location": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Location to create new Datasets in",
			},
			"maximum_bytes_billed": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Max number of bytes that can be billed for a given BigQuery query",
			},
			"gcs_bucket": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URI for a Google Cloud Storage bucket to host Python code executed via Datapro",
			},
			"dataproc_region": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Google Cloud region for PySpark workloads on Dataproc",
			},
			"dataproc_cluster_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Dataproc cluster name for PySpark workloads",
			},
			"application_secret": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The Application Secret for BQ OAuth",
			},
			"application_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The Application ID for BQ OAuth",
			},
			"is_configured_for_oauth": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the connection is configured for OAuth or not",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceBigQueryConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	isActive := d.Get("is_active").(bool)
	projectId := d.Get("project_id").(int)
	name := d.Get("name").(string)
	connectionType := d.Get("type").(string)
	gcpProjectId := d.Get("gcp_project_id").(string)
	timeoutSeconds := d.Get("timeout_seconds").(int)
	privateKeyId := d.Get("private_key_id").(string)
	privateKey := d.Get("private_key").(string)
	clientEmail := d.Get("client_email").(string)
	clientId := d.Get("client_id").(string)
	authUri := d.Get("auth_uri").(string)
	tokenUri := d.Get("token_uri").(string)
	authProviderX509CertUrl := d.Get("auth_provider_x509_cert_url").(string)
	clientX509CertUrl := d.Get("client_x509_cert_url").(string)

	var retriesVal *int
	if d.Get("retries").(int) == 0 {
		retriesVal = nil
	} else {
		retries := d.Get("retries").(int)
		retriesVal = &retries
	}
	var locationVal *string
	if d.Get("location").(string) == "" {
		locationVal = nil
	} else {
		location := d.Get("location").(string)
		locationVal = &location
	}
	var maximumBytesBilledVal *int
	if d.Get("maximum_bytes_billed").(int) == 0 {
		maximumBytesBilledVal = nil
	} else {
		maximumBytesBilled := d.Get("maximum_bytes_billed").(int)
		maximumBytesBilledVal = &maximumBytesBilled
	}
	var gcsBucketVal *string
	if d.Get("gcs_bucket").(string) == "" {
		gcsBucketVal = nil
	} else {
		gcsBucket := d.Get("gcs_bucket").(string)
		gcsBucketVal = &gcsBucket
	}
	var dataprocRegionVal *string
	if d.Get("dataproc_region").(string) == "" {
		dataprocRegionVal = nil
	} else {
		dataprocRegion := d.Get("dataproc_region").(string)
		dataprocRegionVal = &dataprocRegion
	}
	var dataprocClusterNameVal *string
	if d.Get("dataproc_cluster_name").(string) == "" {
		dataprocClusterNameVal = nil
	} else {
		dataprocClusterName := d.Get("dataproc_cluster_name").(string)
		dataprocClusterNameVal = &dataprocClusterName
	}

	applicationSecret := d.Get("application_secret").(string)
	applicationId := d.Get("application_id").(string)

	// scopes := d.Get("scopes").([]string)

	connection, err := c.CreateBigQueryConnection(projectId,
		name,
		connectionType,
		isActive,
		gcpProjectId,
		timeoutSeconds,
		privateKeyId,
		privateKey,
		clientEmail,
		clientId,
		authUri,
		tokenUri,
		authProviderX509CertUrl,
		clientX509CertUrl,
		retriesVal,
		locationVal,
		maximumBytesBilledVal,
		gcsBucketVal,
		dataprocRegionVal,
		dataprocClusterNameVal,
		applicationSecret,
		applicationId)

	// TODO fix scopes
	// scopes)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", connection.ProjectID, dbt_cloud.ID_DELIMITER, *connection.ID))

	resourceBigQueryConnectionRead(ctx, d, m)

	return diags
}

func resourceBigQueryConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	connectionIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	connection, err := c.GetBigQueryConnection(connectionIdString, projectIdString)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("connection_id", connection.ID); err != nil {
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
	if err := d.Set("gcp_project_id", connection.Details.GcpProjectId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("timeout_seconds", connection.Details.TimeoutSeconds); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("private_key_id", connection.Details.PrivateKeyId); err != nil {
		return diag.FromErr(err)
	}
	connection.Details.PrivateKey = d.Get("private_key").(string)
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
	if err := d.Set("gcs_bucket", connection.Details.GcsBucket); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dataproc_region", connection.Details.DataprocRegion); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dataproc_cluster_name", connection.Details.DataprocClusterName); err != nil {
		return diag.FromErr(err)
	}
	connection.Details.ApplicationSecret = d.Get("application_secret").(string)
	connection.Details.ApplicationId = d.Get("application_id").(string)
	if err := d.Set("is_configured_for_oauth", connection.Details.IsConfiguredOAuth); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceBigQueryConnectionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	connectionIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	if d.HasChange("name") ||
		d.HasChange("type") ||
		d.HasChange("gcp_project_id") ||
		d.HasChange("timeout_seconds") ||
		d.HasChange("private_key_id") ||
		d.HasChange("private_key") ||
		d.HasChange("client_email") ||
		d.HasChange("client_id") ||
		d.HasChange("auth_uri") ||
		d.HasChange("token_uri") ||
		d.HasChange("auth_provider_x509_cert_url") ||
		d.HasChange("client_x509_cert_url") ||
		d.HasChange("retries") ||
		d.HasChange("location") ||
		d.HasChange("maximum_bytes_billed") ||
		d.HasChange("gcs_bucket") ||
		d.HasChange("dataproc_region") ||
		d.HasChange("dataproc_cluster_name") ||
		d.HasChange("application_secret") ||
		d.HasChange("application_id") {
		connection, err := c.GetBigQueryConnection(connectionIdString, projectIdString)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("name") {
			name := d.Get("name").(string)
			connection.Name = name
		}
		if d.HasChange("type") {
			connectionType := d.Get("type").(string)
			connection.Type = connectionType
		}
		if d.HasChange("gcp_project_id") {
			gcpProjectId := d.Get("gcp_project_id").(string)
			connection.Details.GcpProjectId = gcpProjectId
		}
		if d.HasChange("timeout_seconds") {
			timeoutSeconds := d.Get("timeout_seconds").(int)
			connection.Details.TimeoutSeconds = timeoutSeconds
		}
		if d.HasChange("private_key_id") {
			privateKeyId := d.Get("private_key_id").(string)
			connection.Details.PrivateKeyId = privateKeyId
		}
		if d.HasChange("private_key") {
			privateKey := d.Get("private_key").(string)
			connection.Details.PrivateKey = privateKey
		}
		if d.HasChange("client_email") {
			clientEmail := d.Get("client_email").(string)
			connection.Details.ClientEmail = clientEmail
		}
		if d.HasChange("client_id") {
			clientId := d.Get("client_id").(string)
			connection.Details.ClientId = clientId
		}
		if d.HasChange("auth_uri") {
			authUri := d.Get("auth_uri").(string)
			connection.Details.AuthUri = authUri
		}
		if d.HasChange("token_uri") {
			tokenUri := d.Get("token_uri").(string)
			connection.Details.TokenUri = tokenUri
		}
		if d.HasChange("auth_provider_x509_cert_url") {
			authProviderX509CertUrl := d.Get("auth_provider_x509_cert_url").(string)
			connection.Details.AuthProviderX509CertUrl = authProviderX509CertUrl
		}
		if d.HasChange("client_x509_cert_url") {
			clientX509CertUrl := d.Get("client_x509_cert_url").(string)
			connection.Details.ClientX509CertUrl = clientX509CertUrl
		}
		if d.HasChange("retries") {
			retries := d.Get("retries").(int)
			if retries == 0 {
				connection.Details.Retries = nil
			} else {
				connection.Details.Retries = &retries
			}
		}
		if d.HasChange("location") {
			location := d.Get("location").(string)
			if location == "" {
				connection.Details.Location = nil
			} else {
				connection.Details.Location = &location
			}
		}
		if d.HasChange("maximum_bytes_billed") {
			maximumBytesBilled := d.Get("maximum_bytes_billed").(int)
			if maximumBytesBilled == 0 {
				connection.Details.MaximumBytesBilled = nil
			} else {
				connection.Details.MaximumBytesBilled = &maximumBytesBilled
			}
		}
		if d.HasChange("gcs_bucket") {
			gcsBucket := d.Get("gcs_bucket").(string)
			if gcsBucket == "" {
				connection.Details.GcsBucket = nil
			} else {
				connection.Details.GcsBucket = &gcsBucket
			}
		}
		if d.HasChange("dataproc_region") {
			dataprocRegion := d.Get("dataproc_region").(string)
			if dataprocRegion == "" {
				connection.Details.DataprocRegion = nil
			} else {
				connection.Details.DataprocRegion = &dataprocRegion
			}
		}
		if d.HasChange("dataproc_cluster_name") {
			dataprocClusterName := d.Get("dataproc_cluster_name").(string)
			if dataprocClusterName == "" {
				connection.Details.DataprocClusterName = nil
			} else {
				connection.Details.DataprocClusterName = &dataprocClusterName
			}
		}
		if d.HasChange("application_secret") {
			applicationSecret := d.Get("application_secret").(string)
			connection.Details.ApplicationSecret = applicationSecret
		}
		if d.HasChange("application_id") {
			applicationId := d.Get("application_id").(string)
			connection.Details.ApplicationId = applicationId
		}

		_, err = c.UpdateBigQueryConnection(connectionIdString, projectIdString, *connection)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceBigQueryConnectionRead(ctx, d, m)
}

func resourceBigQueryConnectionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	return resourceConnectionDelete(ctx, d, m)

}
