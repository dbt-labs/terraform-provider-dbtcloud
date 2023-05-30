resource "dbt_cloud_databricks_credential" "new_credential" {
  project_id  = data.dbt_cloud_project.test_project.project_id
  adapter_id  = 123
  schema      = "my_schema"
  catalog     = "my_catalog"
  token       = "my_secure_token"
}
