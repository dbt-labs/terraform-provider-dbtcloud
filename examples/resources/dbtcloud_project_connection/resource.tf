resource "dbtcloud_project_connection" "dbt_project_connection" {
  project_id    = dbtcloud_project.dbt_project.id
  connection_id = dbtcloud_connection.dbt_connection.connection_id
}