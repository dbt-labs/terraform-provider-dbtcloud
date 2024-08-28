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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

			Those connections are not linked to a specific project and can be linked to environments from different projects by using the ~~~connection_id~~~ field in the ~~~dbtcloud_environment~~~ resource.
			
			All connections types are supported, and the old resources ~~~dbtcloud_connection~~~, ~~~dbtcloud_bigquery_connection~~~ and ~~~dbtcloud_fabric_connection~~~ are now flagged as deprecated and will be removed from the next major version of the provider.`,
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
			"databricks": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Databricks connection configuration",
				Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						Required:    true,
						Description: "The hostname of the Databricks cluster or SQL warehouse.",
					},
					"http_path": schema.StringAttribute{
						Required:    true,
						Description: "The HTTP path of the Databricks cluster or SQL warehouse.",
					},
					"catalog": schema.StringAttribute{
						Optional:    true,
						Description: "Catalog name if Unity Catalog is enabled in your Databricks workspace.",
					},
					"client_id": schema.StringAttribute{
						Optional:    true,
						Description: "Required to enable Databricks OAuth authentication for IDE developers.",
					},
					"client_secret": schema.StringAttribute{
						Optional:    true,
						Description: "Required to enable Databricks OAuth authentication for IDE developers.",
					},
				},
			},
			"redshift": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Redshift connection configuration",
				Attributes: map[string]schema.Attribute{
					"hostname": schema.StringAttribute{
						Required:    true,
						Description: "The hostname of the data warehouse.",
					},
					"port": schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(5432),
						Computed:    true,
						Description: "The port to connect to for this connection. Default=5432",
					},
					"dbname": schema.StringAttribute{
						Optional:    true,
						Description: "The database name for this connection.",
					},
					// for SSH tunnel details
					"ssh_tunnel": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Redshift SSH Tunnel configuration",
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Required:    true,
								Description: "The username to use for the SSH tunnel.",
							},
							"port": schema.Int64Attribute{
								Required:    true,
								Description: "The HTTP port for the SSH tunnel.",
							},
							"hostname": schema.StringAttribute{
								Required:    true,
								Description: "The hostname for the SSH tunnel.",
							},
							"public_key": schema.StringAttribute{
								Computed:    true,
								Description: "The SSH public key generated to allow connecting via SSH tunnel.",
							},
							"id": schema.Int64Attribute{
								Computed:    true,
								Description: "The ID of the SSH tunnel connection.",
							},
						},
					},
				},
			},
			"postgres": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "PostgreSQL connection configuration.",
				Attributes: map[string]schema.Attribute{
					"hostname": schema.StringAttribute{
						Required:    true,
						Description: "The hostname of the database.",
					},
					"port": schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(5432),
						Computed:    true,
						Description: "The port to connect to for this connection. Default=5432",
					},
					"dbname": schema.StringAttribute{
						Optional:    true,
						Description: "The database name for this connection.",
					},
					// for SSH tunnel details
					"ssh_tunnel": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "PostgreSQL SSH Tunnel configuration",
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Required:    true,
								Description: "The username to use for the SSH tunnel.",
							},
							"port": schema.Int64Attribute{
								Required:    true,
								Description: "The HTTP port for the SSH tunnel.",
							},
							"hostname": schema.StringAttribute{
								Required:    true,
								Description: "The hostname for the SSH tunnel.",
							},
							"public_key": schema.StringAttribute{
								Computed:    true,
								Description: "The SSH public key generated to allow connecting via SSH tunnel.",
							},
							"id": schema.Int64Attribute{
								Computed:    true,
								Description: "The ID of the SSH tunnel connection.",
							},
						},
					},
				},
			},
			"fabric": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Microsoft Fabric connection configuration.",
				Attributes: map[string]schema.Attribute{
					"server": schema.StringAttribute{
						Required:    true,
						Description: "The server hostname.",
					},
					"port": schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(1433),
						Computed:    true,
						Description: "The port to connect to for this connection. Default=1433",
					},
					"database": schema.StringAttribute{
						Required:    true,
						Description: "The database to connect to for this connection.",
					},
					"retries": schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(1),
						Computed:    true,
						Description: "The number of automatic times to retry a query before failing. Defaults to 1. Queries with syntax errors will not be retried. This setting can be used to overcome intermittent network issues.",
					},
					"login_timeout": schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(0),
						Computed:    true,
						Description: "The number of seconds used to establish a connection before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
					"query_timeout": schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(0),
						Computed:    true,
						Description: "The number of seconds used to wait for a query before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
				},
			},
			"synapse": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Azure Synapse Analytics connection configuration.",
				Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						Required:    true,
						Description: "The server hostname.",
					},
					"port": schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(1433),
						Computed:    true,
						Description: "The port to connect to for this connection. Default=1433",
					},
					"database": schema.StringAttribute{
						Required:    true,
						Description: "The database to connect to for this connection.",
					},
					"retries": schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(1),
						Computed:    true,
						Description: "The number of automatic times to retry a query before failing. Defaults to 1. Queries with syntax errors will not be retried. This setting can be used to overcome intermittent network issues.",
					},
					"login_timeout": schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(0),
						Computed:    true,
						Description: "The number of seconds used to establish a connection before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
					"query_timeout": schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(0),
						Computed:    true,
						Description: "The number of seconds used to wait for a query before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
				},
			},
			"starburst": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Starburst/Trino connection configuration.",
				Attributes: map[string]schema.Attribute{
					// not too useful now, but should be easy to modify if we support for authentication methods
					"method": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The authentication method. Only LDAP for now.",
						Default:     stringdefault.StaticString("ldap"),
						Validators: []validator.String{
							stringvalidator.OneOf([]string{"ldap"}...),
						},
					},
					"host": schema.StringAttribute{
						Required:    true,
						Description: "The hostname of the account to connect to.",
					},
					"port": schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(443),
						Computed:    true,
						Description: "The port to connect to for this connection. Default=443",
					},
				},
			}, "athena": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Athena connection configuration.",
				Attributes: map[string]schema.Attribute{
					"region_name": schema.StringAttribute{
						Required:    true,
						Description: "AWS region of your Athena instance.",
					},
					"database": schema.StringAttribute{
						Required:    true,
						Description: "Specify the database (data catalog) to build models into (lowercase only).",
					},
					"s3_staging_dir": schema.StringAttribute{
						Required:    true,
						Description: "S3 location to store Athena query results and metadata.",
					},
					"work_group": schema.StringAttribute{
						Optional:    true,
						Description: "Identifier of Athena workgroup.",
					},
					"spark_work_group": schema.StringAttribute{
						Optional:    true,
						Description: "Identifier of Athena Spark workgroup for running Python models.",
					},
					"s3_data_dir": schema.StringAttribute{
						Optional:    true,
						Description: "Prefix for storing tables, if different from the connection's S3 staging directory.",
					},
					"s3_data_naming": schema.StringAttribute{
						Optional:    true,
						Description: "How to generate table paths in the S3 data directory.",
					},
					"s3_tmp_table_dir": schema.StringAttribute{
						Optional:    true,
						Description: "Prefix for storing temporary tables, if different from the connection's S3 data directory.",
					},
					"poll_interval": schema.Int64Attribute{
						Optional:    true,
						Description: "Interval in seconds to use for polling the status of query results in Athena.",
					},
					"num_retries": schema.Int64Attribute{
						Optional:    true,
						Description: "Number of times to retry a failing query.",
					},
					"num_boto3_retries": schema.Int64Attribute{
						Optional:    true,
						Description: "Number of times to retry boto3 requests (e.g. deleting S3 files for materialized tables).",
					},
					"num_iceberg_retries": schema.Int64Attribute{
						Optional:    true,
						Description: "Number of times to retry iceberg commit queries to fix ICEBERG_COMMIT_ERROR.",
					},
				},
			},
			"apache_spark": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Apache Spark connection configuration.",
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						Required:    true,
						Description: "Authentication method for the connection (http or thrift).",
						Validators: []validator.String{
							stringvalidator.OneOf([]string{"http", "thrift"}...),
						},
					},
					"host": schema.StringAttribute{
						Required:    true,
						Description: "Hostname of the connection",
					},
					"port": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Port for the connection. Default=443",
						Default:     int64default.StaticInt64(443),
					},
					"cluster": schema.StringAttribute{
						Required:    true,
						Description: "Spark cluster for the connection",
					},
					"connect_timeout": schema.Int64Attribute{
						Optional:    true,
						Description: "Connection time out in seconds. Default=10",
						Computed:    true,
						Default:     int64default.StaticInt64(10),
					},
					"connect_retries": schema.Int64Attribute{
						Optional:    true,
						Description: "Connection retries. Default=0",
						Computed:    true,
						Default:     int64default.StaticInt64(0),
					},
					"organization": schema.StringAttribute{
						Optional:    true,
						Description: "Organization ID",
					},
					"user": schema.StringAttribute{
						Optional:    true,
						Description: "User",
					},
					"auth": schema.StringAttribute{
						Optional:    true,
						Description: "Auth",
					},
				},
			},
		},
	}
}
