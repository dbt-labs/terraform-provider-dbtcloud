resource "dbtcloud_fabric_connection" "my_fabric_connection" {
  project_id    = dbtcloud_project.dbt_project.id
  name          = "Connection Name"
  server        = "my-server"
  database      = "my-database"
  port          = 1234
  login_timeout = 30
}
