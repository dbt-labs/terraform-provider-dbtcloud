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
    host = "my-databricks-host.cloud.databricks.com"
    http_path = "/sql/my/http/path"
    // optional fields
    catalog = "dbt_catalog"
    client_id = "yourclientid"
    client_secret = "yourclientsecret"
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