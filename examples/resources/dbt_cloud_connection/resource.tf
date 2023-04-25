# Databricks connection
resource "dbt_cloud_connection" "databricks" {
  project_id = 1
  type       = "adapter"
  name       = "Databricks"
  database   = ""
  host_name  = "my-databricks-host.cloud.databricks.com"
  https_path = "/my/path"
  catalog    = "moo"
}

resource "dbt_cloud_connection" "redshift" {
  project_id = 2
  type       = "redshift"
  name       = "My Redshift Warehouse"
  database   = "my-database"
  port       = 5439
  host_name  = "my-redshift-hostname"
}

resource "dbt_cloud_connection" "snowflake" {
  project_id = 3
  type       = "snowflake"
  name       = "My Snowflake warehouse"
  account    = "my-snowflake-account"
  database   = "MY_DATABASE"
  role       = "MY_ROLE"
  warehouse  = "MY_WAREHOUSE"
}
