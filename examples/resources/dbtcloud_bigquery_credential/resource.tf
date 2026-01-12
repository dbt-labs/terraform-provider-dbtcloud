resource "dbtcloud_bigquery_credential" "my_credential" {
  project_id  = dbtcloud_project.dbt_project.id
  dataset     = "my_bq_dataset"
  num_threads = 16
}

# When using a global connection with use_latest_adapter = true,
# provide the connection_id to automatically use the correct adapter version
resource "dbtcloud_bigquery_credential" "my_credential_v1" {
  project_id    = dbtcloud_project.dbt_project.id
  dataset       = "my_bq_dataset"
  num_threads   = 16
  connection_id = dbtcloud_global_connection.my_connection.id
}
