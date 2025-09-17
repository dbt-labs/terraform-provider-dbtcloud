package global_connection

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

	resp.Schema = resource_schema.Schema{
		Description: helper.DocString(
			`This resource can be used to create global connections as introduced in dbt Cloud in August 2024.

			Those connections are not linked to a specific project and can be linked to environments from different projects by using the ~~~connection_id~~~ field in the ~~~dbtcloud_environment~~~ resource.`,
		),
		Attributes: map[string]resource_schema.Attribute{
			"id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "Connection Identifier",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"adapter_version": resource_schema.StringAttribute{
				Computed:    true,
				Description: "Version of the adapter",
			},
			"name": resource_schema.StringAttribute{
				Required:    true,
				Description: "Connection name",
			},
			"is_ssh_tunnel_enabled": resource_schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the connection can use an SSH tunnel",
			},
			"private_link_endpoint_id": resource_schema.StringAttribute{
				Optional:    true,
				Description: "Private Link Endpoint ID. This ID can be found using the `privatelink_endpoint` data source",
			},
			"oauth_configuration_id": resource_schema.Int64Attribute{
				Optional:    true,
				Description: "External OAuth configuration ID (only Snowflake for now)",
			},
			"bigquery": resource_schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]resource_schema.Attribute{
					"gcp_project_id": resource_schema.StringAttribute{
						Required:    true,
						Description: "The GCP project ID to use for the connection",
					},
					"timeout_seconds": resource_schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(300),
						Description: "Timeout in seconds for queries, to be used ONLY for the bigquery_v0 adapter",
					},
					"job_execution_timeout_seconds": resource_schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(300),
						Description: "Timeout in seconds for job execution, to be used for the bigquery_v1 adapter",
					},
					"private_key_id": resource_schema.StringAttribute{
						Required:    true,
						Description: "Private Key ID for the Service Account",
					},
					"private_key": resource_schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "Private Key for the Service Account",
					},
					"client_email": resource_schema.StringAttribute{
						Required:    true,
						Description: "Service Account email",
					},
					"client_id": resource_schema.StringAttribute{
						Required:    true,
						Description: "Client ID of the Service Account",
					},
					"auth_uri": resource_schema.StringAttribute{
						Required:    true,
						Description: "Auth URI for the Service Account",
					},
					"token_uri": resource_schema.StringAttribute{
						Required:    true,
						Description: "Token URI for the Service Account",
					},
					"auth_provider_x509_cert_url": resource_schema.StringAttribute{
						Required:    true,
						Description: "Auth Provider X509 Cert URL for the Service Account",
					},
					"client_x509_cert_url": resource_schema.StringAttribute{
						Required:    true,
						Description: "Client X509 Cert URL for the Service Account",
					},
					"priority": resource_schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.OneOf([]string{"batch", "interactive"}...),
						},
						Description: "The priority with which to execute BigQuery queries (batch or interactive)",
					},
					"retries": resource_schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(1),
						Description: "Number of retries for queries",
					},
					"location": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Location to create new Datasets in",
					},
					"maximum_bytes_billed": resource_schema.Int64Attribute{
						Optional:    true,
						Description: "Max number of bytes that can be billed for a given BigQuery query",
					},
					"execution_project": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Project to bill for query execution",
					},
					"impersonate_service_account": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Service Account to impersonate when running queries",
					},
					"job_retry_deadline_seconds": resource_schema.Int64Attribute{
						Optional:    true,
						Description: "Total number of seconds to wait while retrying the same query",
					},
					"job_creation_timeout_seconds": resource_schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum timeout for the job creation step",
					},
					"application_id": resource_schema.StringAttribute{
						Optional:    true,
						Description: "OAuth Client ID",
						Sensitive:   true,
					},
					"application_secret": resource_schema.StringAttribute{
						Optional:    true,
						Description: "OAuth Client Secret",
						Sensitive:   true,
					},
					"gcs_bucket": resource_schema.StringAttribute{
						Optional:    true,
						Description: "URI for a Google Cloud Storage bucket to host Python code executed via Datapro",
					},
					"dataproc_region": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Google Cloud region for PySpark workloads on Dataproc",
					},
					"dataproc_cluster_name": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Dataproc cluster name for PySpark workloads",
					},
					"scopes": resource_schema.SetAttribute{
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
					"use_legacy_adapter": resource_schema.BoolAttribute{
						Optional:    true,
						Description: "Whether to use the legacy bigquery_v0 adapter. If true, the `timeout_seconds` field will be used",
					},
				},
			},
			// this feels bad, but there is no error/warning when people add extra fields https://github.com/hashicorp/terraform/issues/33570
			"snowflake": resource_schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Snowflake connection configuration",
				Attributes: map[string]resource_schema.Attribute{
					"account": resource_schema.StringAttribute{
						Required:    true,
						Description: "The Snowflake account name",
					},
					"database": resource_schema.StringAttribute{
						Required:    true,
						Description: "The default database for the connection",
					},
					"warehouse": resource_schema.StringAttribute{
						Required:    true,
						Description: "The default Snowflake Warehouse to use for the connection",
					},
					"allow_sso": resource_schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether to allow Snowflake OAuth for the connection. If true, the `oauth_client_id` and `oauth_client_secret` fields must be set",
					},
					// TODO: required if allow_sso is true
					"oauth_client_id": resource_schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "OAuth Client ID. Required to allow OAuth between dbt Cloud and Snowflake",
					},
					"oauth_client_secret": resource_schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "OAuth Client Secret. Required to allow OAuth between dbt Cloud and Snowflake",
					},
					"role": resource_schema.StringAttribute{
						Optional:    true,
						Description: "The Snowflake role to use when running queries on the connection",
					},
					"client_session_keep_alive": resource_schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "If true, the snowflake client will keep connections for longer than the default 4 hours. This is helpful when particularly long-running queries are executing (> 4 hours)",
					},
				},
			},
			"databricks": resource_schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Databricks connection configuration",
				Attributes: map[string]resource_schema.Attribute{
					"host": resource_schema.StringAttribute{
						Required:    true,
						Description: "The hostname of the Databricks cluster or SQL warehouse.",
					},
					"http_path": resource_schema.StringAttribute{
						Required:    true,
						Description: "The HTTP path of the Databricks cluster or SQL warehouse.",
					},
					"catalog": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Catalog name if Unity Catalog is enabled in your Databricks workspace.",
					},
					"client_id": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Required to enable Databricks OAuth authentication for IDE developers.",
					},
					"client_secret": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Required to enable Databricks OAuth authentication for IDE developers.",
					},
				},
			},
			"redshift": resource_schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Redshift connection configuration",
				Attributes: map[string]resource_schema.Attribute{
					"hostname": resource_schema.StringAttribute{
						Required:    true,
						Description: "The hostname of the data warehouse.",
					},
					"port": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(5432),
						Computed:    true,
						Description: "The port to connect to for this connection. Default=5432",
					},
					"dbname": resource_schema.StringAttribute{
						Required:    true,
						Description: "The database name for this connection.",
					},
					// for SSH tunnel details
					"ssh_tunnel": resource_schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Redshift SSH Tunnel configuration",
						Attributes: map[string]resource_schema.Attribute{
							"username": resource_schema.StringAttribute{
								Required:    true,
								Description: "The username to use for the SSH tunnel.",
							},
							"port": resource_schema.Int64Attribute{
								Required:    true,
								Description: "The HTTP port for the SSH tunnel.",
							},
							"hostname": resource_schema.StringAttribute{
								Required:    true,
								Description: "The hostname for the SSH tunnel.",
							},
							"public_key": resource_schema.StringAttribute{
								Computed:    true,
								Description: "The SSH public key generated to allow connecting via SSH tunnel.",
							},
							"id": resource_schema.Int64Attribute{
								Computed:    true,
								Description: "The ID of the SSH tunnel connection.",
							},
						},
					},
				},
			},
			"postgres": resource_schema.SingleNestedAttribute{
				Optional:    true,
				Description: "PostgreSQL connection configuration.",
				Attributes: map[string]resource_schema.Attribute{
					"hostname": resource_schema.StringAttribute{
						Required:    true,
						Description: "The hostname of the database.",
					},
					"port": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(5432),
						Computed:    true,
						Description: "The port to connect to for this connection. Default=5432",
					},
					"dbname": resource_schema.StringAttribute{
						Required:    true,
						Description: "The database name for this connection.",
					},
					// for SSH tunnel details
					"ssh_tunnel": resource_schema.SingleNestedAttribute{
						Optional:    true,
						Description: "PostgreSQL SSH Tunnel configuration",
						Attributes: map[string]resource_schema.Attribute{
							"username": resource_schema.StringAttribute{
								Required:    true,
								Description: "The username to use for the SSH tunnel.",
							},
							"port": resource_schema.Int64Attribute{
								Required:    true,
								Description: "The HTTP port for the SSH tunnel.",
							},
							"hostname": resource_schema.StringAttribute{
								Required:    true,
								Description: "The hostname for the SSH tunnel.",
							},
							"public_key": resource_schema.StringAttribute{
								Computed:    true,
								Description: "The SSH public key generated to allow connecting via SSH tunnel.",
							},
							"id": resource_schema.Int64Attribute{
								Computed:    true,
								Description: "The ID of the SSH tunnel connection.",
							},
						},
					},
				},
			},
			"fabric": resource_schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Microsoft Fabric connection configuration.",
				Attributes: map[string]resource_schema.Attribute{
					"server": resource_schema.StringAttribute{
						Required:    true,
						Description: "The server hostname.",
					},
					"port": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(1433),
						Computed:    true,
						Description: "The port to connect to for this connection. Default=1433",
					},
					"database": resource_schema.StringAttribute{
						Required:    true,
						Description: "The database to connect to for this connection.",
					},
					"retries": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(1),
						Computed:    true,
						Description: "The number of automatic times to retry a query before failing. Defaults to 1. Queries with syntax errors will not be retried. This setting can be used to overcome intermittent network issues.",
					},
					"login_timeout": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(0),
						Computed:    true,
						Description: "The number of seconds used to establish a connection before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
					"query_timeout": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(0),
						Computed:    true,
						Description: "The number of seconds used to wait for a query before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
				},
			},
			"synapse": resource_schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Azure Synapse Analytics connection configuration.",
				Attributes: map[string]resource_schema.Attribute{
					"host": resource_schema.StringAttribute{
						Required:    true,
						Description: "The server hostname.",
					},
					"port": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(1433),
						Computed:    true,
						Description: "The port to connect to for this connection. Default=1433",
					},
					"database": resource_schema.StringAttribute{
						Required:    true,
						Description: "The database to connect to for this connection.",
					},
					"retries": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(1),
						Computed:    true,
						Description: "The number of automatic times to retry a query before failing. Defaults to 1. Queries with syntax errors will not be retried. This setting can be used to overcome intermittent network issues.",
					},
					"login_timeout": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(0),
						Computed:    true,
						Description: "The number of seconds used to establish a connection before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
					"query_timeout": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(0),
						Computed:    true,
						Description: "The number of seconds used to wait for a query before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
				},
			},
			"starburst": resource_schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Starburst/Trino connection configuration.",
				Attributes: map[string]resource_schema.Attribute{
					// not too useful now, but should be easy to modify if we support for authentication methods
					"method": resource_schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The authentication method. Only LDAP for now.",
						Default:     stringdefault.StaticString("ldap"),
						Validators: []validator.String{
							stringvalidator.OneOf([]string{"ldap"}...),
						},
					},
					"host": resource_schema.StringAttribute{
						Required:    true,
						Description: "The hostname of the account to connect to.",
					},
					"port": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(443),
						Computed:    true,
						Description: "The port to connect to for this connection. Default=443",
					},
				},
			},
			"athena": resource_schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Athena connection configuration.",
				Attributes: map[string]resource_schema.Attribute{
					"region_name": resource_schema.StringAttribute{
						Required:    true,
						Description: "AWS region of your Athena instance.",
					},
					"database": resource_schema.StringAttribute{
						Required:    true,
						Description: "Specify the database (data catalog) to build models into (lowercase only).",
					},
					"s3_staging_dir": resource_schema.StringAttribute{
						Required:    true,
						Description: "S3 location to store Athena query results and metadata.",
					},
					"work_group": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Identifier of Athena workgroup.",
					},
					"spark_work_group": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Identifier of Athena Spark workgroup for running Python models.",
					},
					"s3_data_dir": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Prefix for storing tables, if different from the connection's S3 staging directory.",
					},
					"s3_data_naming": resource_schema.StringAttribute{
						Optional:    true,
						Description: "How to generate table paths in the S3 data directory.",
					},
					"s3_tmp_table_dir": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Prefix for storing temporary tables, if different from the connection's S3 data directory.",
					},
					"poll_interval": resource_schema.Int64Attribute{
						Optional:    true,
						Description: "Interval in seconds to use for polling the status of query results in Athena.",
					},
					"num_retries": resource_schema.Int64Attribute{
						Optional:    true,
						Description: "Number of times to retry a failing query.",
					},
					"num_boto3_retries": resource_schema.Int64Attribute{
						Optional:    true,
						Description: "Number of times to retry boto3 requests (e.g. deleting S3 files for materialized tables).",
					},
					"num_iceberg_retries": resource_schema.Int64Attribute{
						Optional:    true,
						Description: "Number of times to retry iceberg commit queries to fix ICEBERG_COMMIT_ERROR.",
					},
				},
			},
			"apache_spark": resource_schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Apache Spark connection configuration.",
				Attributes: map[string]resource_schema.Attribute{
					"method": resource_schema.StringAttribute{
						Required:    true,
						Description: "Authentication method for the connection (http or thrift).",
						Validators: []validator.String{
							stringvalidator.OneOf([]string{"http", "thrift"}...),
						},
					},
					"host": resource_schema.StringAttribute{
						Required:    true,
						Description: "Hostname of the connection",
					},
					"port": resource_schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Port for the connection. Default=443",
						Default:     int64default.StaticInt64(443),
					},
					"cluster": resource_schema.StringAttribute{
						Required:    true,
						Description: "Spark cluster for the connection",
					},
					"connect_timeout": resource_schema.Int64Attribute{
						Optional:    true,
						Description: "Connection time out in seconds. Default=10",
						Computed:    true,
						Default:     int64default.StaticInt64(10),
					},
					"connect_retries": resource_schema.Int64Attribute{
						Optional:    true,
						Description: "Connection retries. Default=0",
						Computed:    true,
						Default:     int64default.StaticInt64(0),
					},
					"organization": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Organization ID",
					},
					"user": resource_schema.StringAttribute{
						Optional:    true,
						Description: "User",
					},
					"auth": resource_schema.StringAttribute{
						Optional:    true,
						Description: "Auth",
					},
				},
			},
			"teradata": resource_schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Teradata connection configuration.",
				Attributes: map[string]resource_schema.Attribute{
					"port": resource_schema.StringAttribute{
						Optional:    true,
						Default:     stringdefault.StaticString("1025"),
						Computed:    true,
						Description: "The port to connect to for this connection. Default=1025",
					},
					//only ANSI supported in cloud at the moment
					"tmode": resource_schema.StringAttribute{
						Required:    true,
						Description: "The transaction mode to use for the connection.",
						Validators: []validator.String{
							stringvalidator.OneOf([]string{"ANSI"}...),
						},
					},
					"host": resource_schema.StringAttribute{
						Required:    true,
						Description: "The hostname of the database.",
					},
					"retries": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(1),
						Computed:    true,
						Description: "The number of automatic times to retry a query before failing. Defaults to 1. Queries with syntax errors will not be retried. This setting can be used to overcome intermittent network issues.",
					},
					"request_timeout": resource_schema.Int64Attribute{
						Optional:    true,
						Default:     int64default.StaticInt64(0),
						Computed:    true,
						Description: "The number of seconds used to establish a connection before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
				},
			},
		},
	}
}

func (r *globalConnectionDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {

	resp.Schema = datasource_schema.Schema{
		Attributes: map[string]datasource_schema.Attribute{
			"id": datasource_schema.Int64Attribute{
				Required:    true,
				Description: "Connection Identifier",
			},
			"adapter_version": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "Version of the adapter",
			},
			"name": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "Connection name",
			},
			"is_ssh_tunnel_enabled": datasource_schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the connection can use an SSH tunnel",
			},
			"private_link_endpoint_id": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "Private Link Endpoint ID. This ID can be found using the `privatelink_endpoint` data source",
			},
			"oauth_configuration_id": datasource_schema.Int64Attribute{
				Computed: true,
			},
			"bigquery": datasource_schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]datasource_schema.Attribute{
					"gcp_project_id": datasource_schema.StringAttribute{
						Required:    true,
						Description: "The GCP project ID to use for the connection",
					},
					"timeout_seconds": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "Timeout in seconds for queries",
					},
					"private_key_id": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Private Key ID for the Service Account",
					},
					"private_key": datasource_schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "Private Key for the Service Account",
					},
					"client_email": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Service Account email",
					},
					"client_id": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Client ID of the Service Account",
					},
					"auth_uri": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Auth URI for the Service Account",
					},
					"token_uri": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Token URI for the Service Account",
					},
					"auth_provider_x509_cert_url": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Auth Provider X509 Cert URL for the Service Account",
					},
					"client_x509_cert_url": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Client X509 Cert URL for the Service Account",
					},
					"priority": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The priority with which to execute BigQuery queries (batch or interactive)",
					},
					"retries": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "Number of retries for queries",
					},
					"location": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Location to create new Datasets in",
					},
					"maximum_bytes_billed": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "Max number of bytes that can be billed for a given BigQuery query",
					},
					"execution_project": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Project to bill for query execution",
					},
					"impersonate_service_account": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Service Account to impersonate when running queries",
					},
					"job_retry_deadline_seconds": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "Total number of seconds to wait while retrying the same query",
					},
					"job_creation_timeout_seconds": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "Maximum timeout for the job creation step",
					},
					"application_id": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "OAuth Client ID",
						Sensitive:   true,
					},
					"application_secret": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "OAuth Client Secret",
						Sensitive:   true,
					},
					"gcs_bucket": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "URI for a Google Cloud Storage bucket to host Python code executed via Datapro",
					},
					"dataproc_region": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Google Cloud region for PySpark workloads on Dataproc",
					},
					"dataproc_cluster_name": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Dataproc cluster name for PySpark workloads",
					},
					"scopes": datasource_schema.SetAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "OAuth scopes for the BigQuery connection",
					},
				},
			},
			// this feels bad, but there is no error/warning when people add extra fields https://github.com/hashicorp/terraform/issues/33570
			"snowflake": datasource_schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Snowflake connection configuration",
				Attributes: map[string]datasource_schema.Attribute{
					"account": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The Snowflake account name",
					},
					"database": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The default database for the connection",
					},
					"warehouse": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The default Snowflake Warehouse to use for the connection",
					},
					"allow_sso": datasource_schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to allow Snowflake OAuth for the connection. If true, the `oauth_client_id` and `oauth_client_secret` fields must be set",
					},
					// TODO: required if allow_sso is true
					"oauth_client_id": datasource_schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "OAuth Client ID. Required to allow OAuth between dbt Cloud and Snowflake",
					},
					"oauth_client_secret": datasource_schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "OAuth Client Secret. Required to allow OAuth between dbt Cloud and Snowflake",
					},
					"role": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The Snowflake role to use when running queries on the connection",
					},
					"client_session_keep_alive": datasource_schema.BoolAttribute{
						Computed:    true,
						Description: "If true, the snowflake client will keep connections for longer than the default 4 hours. This is helpful when particularly long-running queries are executing (> 4 hours)",
					},
				},
			},
			"databricks": datasource_schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Databricks connection configuration",
				Attributes: map[string]datasource_schema.Attribute{
					"host": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The hostname of the Databricks cluster or SQL warehouse.",
					},
					"http_path": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The HTTP path of the Databricks cluster or SQL warehouse.",
					},
					"catalog": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Catalog name if Unity Catalog is enabled in your Databricks workspace.",
					},
					"client_id": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Required to enable Databricks OAuth authentication for IDE developers.",
					},
					"client_secret": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Required to enable Databricks OAuth authentication for IDE developers.",
					},
				},
			},
			"redshift": datasource_schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Redshift connection configuration",
				Attributes: map[string]datasource_schema.Attribute{
					"hostname": datasource_schema.StringAttribute{
						Required:    true,
						Description: "The hostname of the data warehouse.",
					},
					"port": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The port to connect to for this connection. Default=5432",
					},
					"dbname": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The database name for this connection.",
					},
					// for SSH tunnel details
					"ssh_tunnel": datasource_schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Redshift SSH Tunnel configuration",
						Attributes: map[string]datasource_schema.Attribute{
							"username": datasource_schema.StringAttribute{
								Computed:    true,
								Description: "The username to use for the SSH tunnel.",
							},
							"port": datasource_schema.Int64Attribute{
								Computed:    true,
								Description: "The HTTP port for the SSH tunnel.",
							},
							"hostname": datasource_schema.StringAttribute{
								Computed:    true,
								Description: "The hostname for the SSH tunnel.",
							},
							"public_key": datasource_schema.StringAttribute{
								Computed:    true,
								Description: "The SSH public key generated to allow connecting via SSH tunnel.",
							},
							"id": datasource_schema.Int64Attribute{
								Computed:    true,
								Description: "The ID of the SSH tunnel connection.",
							},
						},
					},
				},
			},
			"postgres": datasource_schema.SingleNestedAttribute{
				Computed:    true,
				Description: "PostgreSQL connection configuration.",
				Attributes: map[string]datasource_schema.Attribute{
					"hostname": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The hostname of the database.",
					},
					"port": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The port to connect to for this connection. Default=5432",
					},
					"dbname": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The database name for this connection.",
					},
					// for SSH tunnel details
					"ssh_tunnel": datasource_schema.SingleNestedAttribute{
						Computed:    true,
						Description: "PostgreSQL SSH Tunnel configuration",
						Attributes: map[string]datasource_schema.Attribute{
							"username": datasource_schema.StringAttribute{
								Computed:    true,
								Description: "The username to use for the SSH tunnel.",
							},
							"port": datasource_schema.Int64Attribute{
								Computed:    true,
								Description: "The HTTP port for the SSH tunnel.",
							},
							"hostname": datasource_schema.StringAttribute{
								Computed:    true,
								Description: "The hostname for the SSH tunnel.",
							},
							"public_key": datasource_schema.StringAttribute{
								Computed:    true,
								Description: "The SSH public key generated to allow connecting via SSH tunnel.",
							},
							"id": datasource_schema.Int64Attribute{
								Computed:    true,
								Description: "The ID of the SSH tunnel connection.",
							},
						},
					},
				},
			},
			"fabric": datasource_schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Microsoft Fabric connection configuration.",
				Attributes: map[string]datasource_schema.Attribute{
					"server": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The server hostname.",
					},
					"port": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The port to connect to for this connection. Default=1433",
					},
					"database": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The database to connect to for this connection.",
					},
					"retries": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The number of automatic times to retry a query before failing. Defaults to 1. Queries with syntax errors will not be retried. This setting can be used to overcome intermittent network issues.",
					},
					"login_timeout": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The number of seconds used to establish a connection before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
					"query_timeout": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The number of seconds used to wait for a query before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
				},
			},
			"synapse": datasource_schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Azure Synapse Analytics connection configuration.",
				Attributes: map[string]datasource_schema.Attribute{
					"host": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The server hostname.",
					},
					"port": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The port to connect to for this connection. Default=1433",
					},
					"database": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The database to connect to for this connection.",
					},
					"retries": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The number of automatic times to retry a query before failing. Defaults to 1. Queries with syntax errors will not be retried. This setting can be used to overcome intermittent network issues.",
					},
					"login_timeout": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The number of seconds used to establish a connection before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
					"query_timeout": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The number of seconds used to wait for a query before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
				},
			},
			"starburst": datasource_schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Starburst/Trino connection configuration.",
				Attributes: map[string]datasource_schema.Attribute{
					// not too useful now, but should be easy to modify if we support for authentication methods
					"method": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The authentication method. Only LDAP for now.",
					},
					"host": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The hostname of the account to connect to.",
					},
					"port": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The port to connect to for this connection. Default=443",
					},
				},
			},
			"athena": datasource_schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Athena connection configuration.",
				Attributes: map[string]datasource_schema.Attribute{
					"region_name": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "AWS region of your Athena instance.",
					},
					"database": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Specify the database (data catalog) to build models into (lowercase only).",
					},
					"s3_staging_dir": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "S3 location to store Athena query results and metadata.",
					},
					"work_group": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Identifier of Athena workgroup.",
					},
					"spark_work_group": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Identifier of Athena Spark workgroup for running Python models.",
					},
					"s3_data_dir": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Prefix for storing tables, if different from the connection's S3 staging directory.",
					},
					"s3_data_naming": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "How to generate table paths in the S3 data directory.",
					},
					"s3_tmp_table_dir": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Prefix for storing temporary tables, if different from the connection's S3 data directory.",
					},
					"poll_interval": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "Interval in seconds to use for polling the status of query results in Athena.",
					},
					"num_retries": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "Number of times to retry a failing query.",
					},
					"num_boto3_retries": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "Number of times to retry boto3 requests (e.g. deleting S3 files for materialized tables).",
					},
					"num_iceberg_retries": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "Number of times to retry iceberg commit queries to fix ICEBERG_COMMIT_ERROR.",
					},
				},
			},
			"apache_spark": datasource_schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Apache Spark connection configuration.",
				Attributes: map[string]datasource_schema.Attribute{
					"method": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Authentication method for the connection (http or thrift).",
					},
					"host": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Hostname of the connection",
					},
					"port": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "Port for the connection. Default=443",
					},
					"cluster": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Spark cluster for the connection",
					},
					"connect_timeout": datasource_schema.Int64Attribute{
						Description: "Connection time out in seconds. Default=10",
						Computed:    true,
					},
					"connect_retries": datasource_schema.Int64Attribute{
						Description: "Connection retries. Default=0",
						Computed:    true,
					},
					"organization": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Organization ID",
					},
					"user": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "User",
					},
					"auth": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Auth",
					},
				},
			},
			"teradata": datasource_schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Teradata connection configuration.",
				Attributes: map[string]datasource_schema.Attribute{
					"host": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The hostname of the database.",
					},
					"port": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The port to connect to for this connection. Default=1025",
					},
					//only ANSI supported in cloud at the moment
					"tmode": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "The transaction mode to use for the connection.",
						Validators: []validator.String{
							stringvalidator.OneOf([]string{"ANSI"}...),
						},
					},
					"retries": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The number of automatic times to retry a query before failing. Defaults to 1. Queries with syntax errors will not be retried. This setting can be used to overcome intermittent network issues.",
					},
					"request_timeout": resource_schema.Int64Attribute{
						Computed:    true,
						Description: "The number of seconds used to establish a connection before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
					},
				},
			},
		},
	}
}

func (r *globalConnectionsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {

	resp.Schema = datasource_schema.Schema{
		Description: "All the connections created on the account with some summary information, like their name, type, when they were created/updated and the number of environments using them.",
		Attributes: map[string]datasource_schema.Attribute{
			"connections": datasource_schema.SetNestedAttribute{
				Computed:    true,
				Description: "A set of all the connections",
				NestedObject: datasource_schema.NestedAttributeObject{
					Attributes: map[string]datasource_schema.Attribute{
						"id": datasource_schema.Int64Attribute{
							Computed:    true,
							Description: "Connection Identifier",
						},
						"created_at": datasource_schema.StringAttribute{
							Computed:    true,
							Description: "When the connection was created",
						},
						"updated_at": datasource_schema.StringAttribute{
							Computed:    true,
							Description: "When the connection was updated",
						},
						"name": datasource_schema.StringAttribute{
							Computed:    true,
							Description: "Connection name",
						},
						"adapter_version": datasource_schema.StringAttribute{
							Computed:    true,
							Description: "Type of adapter used for the connection",
						},
						"private_link_endpoint_id": datasource_schema.StringAttribute{
							Computed:    true,
							Description: "Private Link Endpoint ID.",
						},
						"is_ssh_tunnel_enabled": datasource_schema.BoolAttribute{
							Computed: true,
						},
						"oauth_configuration_id": datasource_schema.Int64Attribute{
							Computed: true,
						},
						"environment__count": datasource_schema.Int64Attribute{
							Computed:    true,
							Description: "Number of environments using this connection",
						},
					},
				},
			},
		},
	}
}
