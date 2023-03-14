# Databricks connection
resource "dbt_cloud_connection" "databricks" {
  project_id = 1
  name       = "Databricks"
  database   = ""

  host_name  = "my-databricks-host.cloud.databricks.com"
  https_path = "/my/path"
  catalog    = "moo"
}
