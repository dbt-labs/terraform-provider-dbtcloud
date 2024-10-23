# when using the Databricks adapter with a new `dbtcloud_global_connection`
# we don't provide an `adapter_id`
resource "dbtcloud_databricks_credential" "my_databricks_cred" {
  project_id   = dbtcloud_project.dbt_project.id
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "databricks"
}

# when using the Databricks adapter with a legacy `dbtcloud_connection`
# we provide an `adapter_id`
resource "dbtcloud_databricks_credential" "my_databricks_cred" {
  project_id   = dbtcloud_project.dbt_project.id
  adapter_id   = dbtcloud_connection.my_databricks_connection.adapter_id
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "databricks"
}

# when using the Spark adapter
resource "dbtcloud_databricks_credential" "my_spark_cred" {
  project_id   = dbtcloud_project.dbt_project.id
  adapter_id   = dbtcloud_connection.my_databricks_connection.adapter_id
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "spark"
}