resource "dbtcloud_global_connection" "apache_spark" {
  name = "My Apache Spark connection"
  apache_spark = {
    method  = "http"
    host    = "my-spark-host.com"
    cluster = "my-cluster"
    // example of optional fields
    connect_timeout = 100
  }
}

resource "dbtcloud_global_connection" "athena" {
  name = "My Athena connection"
  athena = {
    region_name    = "us-east-1"
    database       = "mydatabase"
    s3_staging_dir = "my_dir"
    // example of optional fields
    work_group = "my_work_group"
  }
}

resource "dbtcloud_global_connection" "bigquery" {
  name = "My BigQuery connection"
  bigquery = {
    gcp_project_id              = "my-gcp-project-id"
    timeout_seconds             = 1000
    private_key_id              = "my-private-key-id"
    private_key                 = "ABCDEFGHIJKL"
    client_email                = "my_client_email"
    client_id                   = "my_client_id"
    auth_uri                    = "my_auth_uri"
    token_uri                   = "my_token_uri"
    auth_provider_x509_cert_url = "my_auth_provider_x509_cert_url"
    client_x509_cert_url        = "my_client_x509_cert_url"
    application_id              = "oauth_application_id"
    application_secret          = "oauth_secret_id"
  }
}

resource "dbtcloud_global_connection" "databricks" {
  name = "My Databricks connection"
  databricks = {
    host      = "my-databricks-host.cloud.databricks.com"
    http_path = "/sql/my/http/path"
    // optional fields
    catalog       = "dbt_catalog"
    client_id     = "yourclientid"
    client_secret = "yourclientsecret"
  }
}

resource "dbtcloud_global_connection" "fabric" {
  name = "My Fabric connection"
  fabric = {
    server   = "my-fabric-server.com"
    database = "mydb"
    // optional fields
    port          = 1234
    retries       = 3
    login_timeout = 60
    query_timeout = 3600
  }
}

resource "dbtcloud_global_connection" "postgres" {
  name = "My PostgreSQL connection"
  postgres = {
    hostname = "my-postgresql-server.com"
    port     = 5432
    // optional fields
    dbname = "my_database"
    // it is possible to set settings to connect via SSH Tunnel as well
  }
}

resource "dbtcloud_global_connection" "redshift" {
  name = "My Redshift connection"
  redshift = {
    hostname = "my-redshift-connection.com"
    port     = 5432
    // optional fields
    dbname = "my_database"
    // it is possible to set settings to connect via SSH Tunnel as well
  }
}

resource "dbtcloud_global_connection" "snowflake" {
  name = "My Snowflake connection"
  // we can set Privatelink if needed
  private_link_endpoint_id = data.dbtcloud_privatelink_endpoint.my_private_link.id
  snowflake = {
    account                   = "my-snowflake-account"
    database                  = "MY_DATABASE"
    warehouse                 = "MY_WAREHOUSE"
    client_session_keep_alive = false
    allow_sso                 = true
    oauth_client_id           = "yourclientid"
    oauth_client_secret       = "yourclientsecret"
  }
}

resource "dbtcloud_global_connection" "starburst" {
  name = "My Starburst connection"
  starburst = {
    host     = "my-starburst-host.com"
    database = "mydb"
  }
}

resource "dbtcloud_global_connection" "synapse" {
  name = "My Synapse connection"
  synapse = {
    host     = "my-synapse-server.com"
    database = "mydb"
    // optional fields
    port          = 1234
    retries       = 3
    login_timeout = 60
    query_timeout = 3600
  }
}

resource "dbtcloud_global_connection" "teradata" {
  name = "My Teradata connection"

  teradata = {
    host       = "teradata.example.com"
    port       		= 1025
    tmode	   		= "ANSI"
    retries	   		= 3
    request_timeout = 3000
  }
}