// use dbt_cloud_databricks_credential instead of dbtcloud_databricks_credential for the legacy resource names
// legacy names will be removed from 0.3 onwards

# when using the Databricks adapter
resource "dbtcloud_databricks_credential" "databricks_cred" {
  project_id   = dbtcloud_project.my_project.id
  adapter_id   = 123
  target_name  = "prod"
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "databricks"
}

# when using the Spark adapter
resource "dbtcloud_databricks_credential" "spark_cred" {
  project_id   = dbtcloud_project.my_other_project.id
  adapter_id   = 456
  target_name  = "prod"
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "spark"
}