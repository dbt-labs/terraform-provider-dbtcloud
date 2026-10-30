// Write-only secrets (`*_wo`) are not stored in state and need Terraform >= 1.11.
// Change the matching `_wo_version` to send a new value.
variable "snowflake_private_key" {
  type      = string
  ephemeral = true
}

variable "snowflake_private_key_passphrase" {
  type      = string
  ephemeral = true
}

variable "databricks_token" {
  type      = string
  ephemeral = true
}

variable "bigquery_private_key" {
  type      = string
  ephemeral = true
}

variable "redshift_password" {
  type      = string
  ephemeral = true
}

// The adapter block must match the adapter of the connection
resource "dbtcloud_analytics_credential" "snowflake" {
  connection_id = dbtcloud_global_connection.snowflake.id

  snowflake = {
    auth_type                         = "keypair"
    user                              = "ANALYTICS_USER"
    role                              = "ANALYTICS_ROLE"
    warehouse                         = "ANALYTICS_WH"
    private_key_wo                    = var.snowflake_private_key
    private_key_wo_version            = 1
    private_key_passphrase_wo         = var.snowflake_private_key_passphrase
    private_key_passphrase_wo_version = 1
  }
}

resource "dbtcloud_analytics_credential" "databricks" {
  connection_id = dbtcloud_global_connection.databricks.id

  databricks = {
    token_wo         = var.databricks_token
    token_wo_version = 1
    catalog          = "analytics"
  }
}

resource "dbtcloud_analytics_credential" "redshift" {
  connection_id = dbtcloud_global_connection.redshift.id

  redshift = {
    username            = "analytics_user"
    password_wo         = var.redshift_password
    password_wo_version = 1
  }
}

// A secret can also be a plain value: it works on any Terraform version, is marked sensitive
// and is stored in the state. It cannot come from an ephemeral variable.
variable "redshift_plain_password" {
  type      = string
  sensitive = true
}

resource "dbtcloud_analytics_credential" "redshift_plain_password" {
  connection_id = dbtcloud_global_connection.redshift_other.id

  redshift = {
    username = "analytics_user"
    password = var.redshift_plain_password
  }
}

// BigQuery with a service account (`auth_type` defaults to "service-account-json")
resource "dbtcloud_analytics_credential" "bigquery" {
  connection_id = dbtcloud_global_connection.bigquery.id

  bigquery = {
    private_key_id              = "my-private-key-id"
    private_key_wo              = var.bigquery_private_key
    private_key_wo_version      = 1
    client_email                = "analytics@my-project.iam.gserviceaccount.com"
    client_id                   = "123456789"
    auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
    token_uri                   = "https://oauth2.googleapis.com/token"
    auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
    client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/analytics%40my-project.iam.gserviceaccount.com"
    execution_project           = "my-billing-project"
  }
}

// BigQuery with workload identity federation (latest adapter only, no secret)
resource "dbtcloud_analytics_credential" "bigquery_wif" {
  connection_id = dbtcloud_global_connection.bigquery_latest.id

  bigquery = {
    auth_type                   = "external-oauth-wif"
    workload_pool_provider_path = "//iam.googleapis.com/projects/123/locations/global/workloadIdentityPools/my-pool/providers/my-provider"
  }
}
