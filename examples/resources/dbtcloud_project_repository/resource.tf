resource "dbtcloud_project_repository" "dbt_project_repository" {
  project_id    = dbtcloud_project.dbt_project.id
  repository_id = dbtcloud_repository.dbt_repository.repository_id
}