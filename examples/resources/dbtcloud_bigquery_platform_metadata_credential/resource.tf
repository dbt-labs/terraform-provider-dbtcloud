// Service account key, with the write-only attribute (not stored in state, requires Terraform >= 1.11).
// The connection can use bigquery_v0 or bigquery_v1.
variable "bigquery_metadata_private_key" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_bigquery_platform_metadata_credential" "example" {
  connection_id = dbtcloud_global_connection.bigquery.id

  catalog_ingestion_enabled = true
  cost_optimization_enabled = false
  cost_insights_enabled     = true

  // to rotate the key, set a new private_key_wo and bump the version
  // (also bump it when private_key_id changes)
  private_key_wo         = var.bigquery_metadata_private_key
  private_key_wo_version = 1

  private_key_id              = "my-private-key-id"
  client_email                = "my-service-account@my-project.iam.gserviceaccount.com"
  client_id                   = "123456789012345678901"
  auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
  token_uri                   = "https://oauth2.googleapis.com/token"
  auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
  client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/my-service-account%40my-project.iam.gserviceaccount.com"
}

// Service account key read from the JSON key file, with the classic sensitive attribute (stored in state)
locals {
  bigquery_key = jsondecode(file("service-account.json"))
}

resource "dbtcloud_bigquery_platform_metadata_credential" "example_from_file" {
  connection_id = dbtcloud_global_connection.bigquery.id

  cost_insights_enabled = true

  private_key                 = local.bigquery_key.private_key
  private_key_id              = local.bigquery_key.private_key_id
  client_email                = local.bigquery_key.client_email
  client_id                   = local.bigquery_key.client_id
  auth_uri                    = local.bigquery_key.auth_uri
  token_uri                   = local.bigquery_key.token_uri
  auth_provider_x509_cert_url = local.bigquery_key.auth_provider_x509_cert_url
  client_x509_cert_url        = local.bigquery_key.client_x509_cert_url
}

// One credential per connection, for example across many projects
resource "dbtcloud_bigquery_platform_metadata_credential" "per_connection" {
  for_each = dbtcloud_global_connection.bigquery_per_project

  connection_id         = each.value.id
  cost_insights_enabled = true

  private_key                 = local.bigquery_key.private_key
  private_key_id              = local.bigquery_key.private_key_id
  client_email                = local.bigquery_key.client_email
  client_id                   = local.bigquery_key.client_id
  auth_uri                    = local.bigquery_key.auth_uri
  token_uri                   = local.bigquery_key.token_uri
  auth_provider_x509_cert_url = local.bigquery_key.auth_provider_x509_cert_url
  client_x509_cert_url        = local.bigquery_key.client_x509_cert_url
}

// Workload Identity Federation. Requires a bigquery_v1 connection (use_latest_adapter = true).
// service_account_impersonation_url is filled in from the connection's impersonate_service_account
// when it is not set.
resource "dbtcloud_bigquery_platform_metadata_credential" "example_wif" {
  connection_id = dbtcloud_global_connection.bigquery_v1.id

  cost_insights_enabled = true

  auth_type                   = "external-oauth-wif"
  workload_pool_provider_path = "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/my-pool/providers/my-provider"
}
