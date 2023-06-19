// use dbt_cloud_connection instead of dbtcloud_connection for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_connection" "databricks" {
  project_id = 1
  type       = "adapter"
  name       = "Databricks"
  database   = ""
  host_name  = "my-databricks-host.cloud.databricks.com"
  https_path = "/my/path"
  catalog    = "moo"
}

resource "dbtcloud_connection" "redshift" {
  project_id = 2
  type       = "redshift"
  name       = "My Redshift Warehouse"
  database   = "my-database"
  port       = 5439
  host_name  = "my-redshift-hostname"
}

resource "dbtcloud_connection" "snowflake" {
  project_id = 3
  type       = "snowflake"
  name       = "My Snowflake warehouse"
  account    = "my-snowflake-account"
  database   = "MY_DATABASE"
  role       = "MY_ROLE"
  warehouse  = "MY_WAREHOUSE"
}