// use dbt_cloud_project instead of dbtcloud_project for the legacy resource names
// legacy names will be removed from 0.3 onwards

data "dbtcloud_project" "test_project" {
  project_id = var.dbt_cloud_project_id
}
