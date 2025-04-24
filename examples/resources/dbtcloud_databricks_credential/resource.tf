resource "dbtcloud_databricks_credential" "my_databricks_cred" {
  project_id   = dbtcloud_project.dbt_project.id
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "databricks"
}