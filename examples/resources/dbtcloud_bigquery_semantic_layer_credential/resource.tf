resource "dbtcloud_bigquery_semantic_layer_credential" "example" {
  configuration = {
    project_id      = var.project_id
	name            = "BigQuery SL Credential"
	adapter_version = "bigquery_v0"
  }
  credential = {
  	project_id = var.project_id
	is_active = true
    num_threads = var.num_threads
	dataset = "test_dataset"
  }
  private_key_id = var.private_key_id
  private_key = var.private_key
  client_email = var.client_email
  client_id = var.client_id
  auth_uri = var.auth_uri
  token_uri = var.token_uri
  auth_provider_x509_cert_url = var.auth_provider_x509_cert_url
  client_x509_cert_url = var.client_x509_cert_url
  
}