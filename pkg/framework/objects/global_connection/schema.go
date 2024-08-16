package global_connection

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *globalConnectionResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {

	resp.Schema = schema.Schema{
		Description: helper.DocString(
			`This resource can be used to create global connections as introduced in dbt Cloud in August 2024.

			Those connections are not linked to a project and can be linked to environments from different projects by using the ~~~connection_id~~~ field in the ~~~dbtcloud_environment~~~ resource.
			
			For now, only BigQuery and Snowflake connections are supported and the other Data Warehouses can continue using the existing resources ~~~dbtcloud_connection~~~ and ~~~dbtcloud_fabric_connection~~~ , 
			but all Data Warehouses will soon be supported under this resource and the other ones will be deprecated in the future.`,
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Connection Identifier",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"adapter_version": schema.StringAttribute{
				Computed:    true,
				Description: "Version of the adapter",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Connection name",
			},
			"is_ssh_tunnel_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the connection can use an SSH tunnel",
			},
			"private_link_endpoint_id": schema.StringAttribute{
				Optional:    true,
				Description: "Private Link Endpoint ID. This ID can be found using the `privatelink_endpoint` data source",
			},
			"oauth_configuration_id": schema.Int64Attribute{
				Computed: true,
			},
			"bigquery": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"gcp_project_id": schema.StringAttribute{
						Required:    true,
						Description: "The GCP project ID to use for the connection",
					},
					"timeout_seconds": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(300),
						Description: "Timeout in seconds for queries",
					},
					"private_key_id": schema.StringAttribute{
						Required:    true,
						Description: "Private Key ID for the Service Account",
					},
					"private_key": schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "Private Key for the Service Account",
					},
					"client_email": schema.StringAttribute{
						Required:    true,
						Description: "Service Account email",
					},
					"client_id": schema.StringAttribute{
						Required:    true,
						Description: "Client ID of the Service Account",
					},
					"auth_uri": schema.StringAttribute{
						Required:    true,
						Description: "Auth URI for the Service Account",
					},
					"token_uri": schema.StringAttribute{
						Required:    true,
						Description: "Token URI for the Service Account",
					},
					"auth_provider_x509_cert_url": schema.StringAttribute{
						Required:    true,
						Description: "Auth Provider X509 Cert URL for the Service Account",
					},
					"client_x509_cert_url": schema.StringAttribute{
						Required:    true,
						Description: "Client X509 Cert URL for the Service Account",
					},
					"priority": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.OneOf([]string{"batch", "interactive"}...),
						},
						Description: "The priority with which to execute BigQuery queries (batch or interactive)",
					},
					"retries": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(1),
						Description: "Number of retries for queries",
					},
					"location": schema.StringAttribute{
						Optional:    true,
						Description: "Location to create new Datasets in",
					},
					"maximum_bytes_billed": schema.Int64Attribute{
						Optional:    true,
						Description: "Max number of bytes that can be billed for a given BigQuery query",
					},
					"execution_project": schema.StringAttribute{
						Optional:    true,
						Description: "Project to bill for query execution",
					},
					"impersonate_service_account": schema.StringAttribute{
						Optional:    true,
						Description: "Service Account to impersonate when running queries",
					},
					"job_retry_deadline_seconds": schema.Int64Attribute{
						Optional:    true,
						Description: "Total number of seconds to wait while retrying the same query",
					},
					"job_creation_timeout_seconds": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum timeout for the job creation step",
					},
					"application_id": schema.StringAttribute{
						Optional:    true,
						Description: "OAuth Client ID",
						Sensitive:   true,
					},
					"application_secret": schema.StringAttribute{
						Optional:    true,
						Description: "OAuth Client Secret",
						Sensitive:   true,
					},
					"gcs_bucket": schema.StringAttribute{
						Optional:    true,
						Description: "URI for a Google Cloud Storage bucket to host Python code executed via Datapro",
					},
					"dataproc_region": schema.StringAttribute{
						Optional:    true,
						Description: "Google Cloud region for PySpark workloads on Dataproc",
					},
					"dataproc_cluster_name": schema.StringAttribute{
						Optional:    true,
						Description: "Dataproc cluster name for PySpark workloads",
					},
					"scopes": schema.SetAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Default: setdefault.StaticValue(
							types.SetValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("https://www.googleapis.com/auth/bigquery"),
									types.StringValue(
										"https://www.googleapis.com/auth/cloud-platform",
									),
									types.StringValue("https://www.googleapis.com/auth/drive"),
								},
							),
						),
						Description: "OAuth scopes for the BigQuery connection",
					},
				},
			},
			// this feels bad, but there is no error/warning when people add extra fields https://github.com/hashicorp/terraform/issues/33570
			"snowflake": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Snowflake connection configuration",
				Attributes: map[string]schema.Attribute{
					"account": schema.StringAttribute{
						Required:    true,
						Description: "The Snowflake account name",
					},
					"database": schema.StringAttribute{
						Required:    true,
						Description: "The default database for the connection",
					},
					"warehouse": schema.StringAttribute{
						Required:    true,
						Description: "The default Snowflake Warehouse to use for the connection",
					},
					"allow_sso": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether to allow Snowflake OAuth for the connection. If true, the `oauth_client_id` and `oauth_client_secret` fields must be set",
					},
					// TODO: required if allow_sso is true
					"oauth_client_id": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "OAuth Client ID. Required to allow OAuth between dbt Cloud and Snowflake",
					},
					"oauth_client_secret": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "OAuth Client Secret. Required to allow OAuth between dbt Cloud and Snowflake",
					},
					"role": schema.StringAttribute{
						Optional:    true,
						Description: "The Snowflake role to use when running queries on the connection",
					},
					"client_session_keep_alive": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "If true, the snowflake client will keep connections for longer than the default 4 hours. This is helpful when particularly long-running queries are executing (> 4 hours)",
					},
				},
			},
		},
	}
}
