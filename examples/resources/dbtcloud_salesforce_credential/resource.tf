# Create a Salesforce credential for dbt Cloud using JWT Bearer Flow authentication
resource "dbtcloud_salesforce_credential" "my_salesforce_cred" {
  project_id  = dbtcloud_project.dbt_project.id
  username    = "user@example.com"
  client_id   = "your-oauth-client-id"
  private_key = "private-key value"
  target_name = "default"
  num_threads = 6
}
