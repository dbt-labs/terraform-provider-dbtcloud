resource "dbtcloud_bigquery_credential" "my_credential" {
  project_id  = dbtcloud_project.dbt_project.id
  dataset     = "my_bq_dataset"
  num_threads = 16
}
