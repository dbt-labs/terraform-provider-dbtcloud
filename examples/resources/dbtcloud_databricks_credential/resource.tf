// NOTE for customers using the LEGACY dbt_cloud provider:
// use dbt_cloud_databricks_credential instead of dbtcloud_databricks_credential for the legacy resource names
// legacy names will be removed from 0.3 onwards

# when using the Databricks adapter
resource "dbtcloud_databricks_credential" "my_databricks_cred" {
  project_id   = dbtcloud_project.dbt_project.id
  adapter_id   = 123
  target_name  = "prod"
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "databricks"
}

# when using the Spark adapter
resource "dbtcloud_databricks_credential" "my_spark_cred" {
  project_id   = dbtcloud_project.dbt_project.id
  adapter_id   = 456
  target_name  = "prod"
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "spark"
}