// NOTE for customers using the LEGACY dbt_cloud provider:
// use dbt_cloud_project_repository instead of dbtcloud_project_repository for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_project_repository" "dbt_project_repository" {
  project_id    = dbtcloud_project.dbt_project.id
  repository_id = dbtcloud_repository.dbt_repository.repository_id
}