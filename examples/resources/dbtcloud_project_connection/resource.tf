// use dbt_cloud_project_connection instead of dbtcloud_project_connection for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_project_connection" "my_project_connection" {
  project_id    = dbtcloud_project.my_project.id
  connection_id = dbtcloud_connection.my_connection.connection_id
}