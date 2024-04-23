# when using the Databricks adapter
resource "dbtcloud_databricks_credential" "my_databricks_cred" {
  project_id   = dbtcloud_project.dbt_project.id
  adapter_id   = dbtcloud_connection.my_databricks_connection.adapter_id
  target_name  = "prod"
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "databricks"
}

# when using the Spark adapter
resource "dbtcloud_databricks_credential" "my_spark_cred" {
  project_id   = dbtcloud_project.dbt_project.id
  adapter_id   = dbtcloud_connection.my_databricks_connection.adapter_id
  target_name  = "prod"
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "spark"
}