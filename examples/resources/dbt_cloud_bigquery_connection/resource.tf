resource "dbt_cloud_bigquery_connection" "test_connection" {
  project_id                  = dbt_cloud_project.test_project.id
  name                        = "Project Name"
  type                        = "bigquery"
  is_active                   = true
  gcp_project_id              = "my-gcp-project-id"
  timeout_seconds             = 100
  private_key_id              = "my-private-key-id"
  private_key                 = "ABCDEFGHIJKL"
  client_email                = "my_client_email"
  client_id                   = "my_client_di"
  auth_uri                    = "my_auth_uri"
  token_uri                   = "my_token_uri"
  auth_provider_x509_cert_url = "my_auth_provider_x509_cert_url"
  client_x509_cert_url        = "my_client_x509_cert_url"
  retries                     = 3
}

# it is also possible to set the connection to use OAuth by filling in `application_id` and `application_secret`
resource "dbt_cloud_bigquery_connection" "test_connection_with_oauth" {
  project_id                  = dbt_cloud_project.test_project.id
  name                        = "Project Name"
  type                        = "bigquery"
  is_active                   = true
  gcp_project_id              = "my-gcp-project-id"
  timeout_seconds             = 100
  private_key_id              = "my-private-key-id"
  private_key                 = "ABCDEFGHIJKL"
  client_email                = "my_client_email"
  client_id                   = "my_client_di"
  auth_uri                    = "my_auth_uri"
  token_uri                   = "my_token_uri"
  auth_provider_x509_cert_url = "my_auth_provider_x509_cert_url"
  client_x509_cert_url        = "my_client_x509_cert_url"
  retries                     = 3
  appplication_id             = "oauth_application_id"
  application_secret          = "oauth_secret_id"
}
