resource "dbt_cloud_databricks_credential" "new_credential_dx" {
  project_id   = data.dbt_cloud_project.test_project_1.project_id
  adapter_id   = 123
  schema       = "my_schema"
  catalog      = "my_catalog"
  token        = "my_secure_token"
  adapter_type = "databricks"
}

resource "dbt_cloud_databricks_credential" "new_credential_spark" {
  project_id   = data.dbt_cloud_project.test_project_2.project_id
  adapter_id   = 456
  schema       = "my_schema"
  token        = "my_secure_token"
  adapter_type = "spark"
}
