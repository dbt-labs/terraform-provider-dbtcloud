// use dbt_cloud_project instead of dbtcloud_project for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_project" "dbt_project" {
  name = "Analytics"
}

resource "dbtcloud_project" "dbt_project_with_subdir" {
  name                     = "Analytics in Subdir"
  dbt_project_subdirectory = "/path"
}