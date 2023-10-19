// NOTE for customers using the LEGACY dbt_cloud provider:
// use dbt_cloud_connection instead of dbtcloud_connection for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_connection" "databricks" {
  project_id = dbtcloud_project.dbt_project.id
  type       = "adapter"
  name       = "Databricks"
  database   = "" // currenyly need to be empty for databricks
  host_name  = "my-databricks-host.cloud.databricks.com"
  http_path  = "/my/path"
  catalog    = "moo"
}

resource "dbtcloud_connection" "redshift" {
  project_id = dbtcloud_project.dbt_project.id
  type       = "redshift"
  name       = "My Redshift Warehouse"
  database   = "my-database"
  port       = 5439
  host_name  = "my-redshift-hostname"
}

resource "dbtcloud_connection" "snowflake" {
  project_id = dbtcloud_project.dbt_project.id
  type       = "snowflake"
  name       = "My Snowflake warehouse"
  account    = "my-snowflake-account"
  database   = "MY_DATABASE"
  role       = "MY_ROLE"
  warehouse  = "MY_WAREHOUSE"
}