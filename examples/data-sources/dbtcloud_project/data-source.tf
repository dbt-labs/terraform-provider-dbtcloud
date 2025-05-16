// use dbt_cloud_project instead of dbtcloud_project for the legacy resource names
// legacy names will be removed from 0.3 onwards

// projects data sources can use the project_id parameter (preferred uniqueness is ensured)
data "dbtcloud_project" "test_project" {
  project_id = var.dbt_cloud_project_id
}

// or they can use project names
// the provider will raise an error if more than one project is found with the same name
data "dbtcloud_project" "test_project" {
  name = "My project name"
}
