// NOTE for customers using the LEGACY dbt_cloud provider:
// use dbt_cloud_bigquery_credential instead of dbtcloud_bigquery_credential for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_bigquery_credential" "my_credential" {
  project_id  = dbtcloud_project.dbt_project.id
  dataset     = "my_bq_dataset"
  num_threads = 16
}
