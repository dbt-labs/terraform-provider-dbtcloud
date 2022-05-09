resource "dbt_cloud_snowflake_credential" "new_credential" {
  project_id  = data.dbt_cloud_project.test_project.project_id
  auth_type   = "password"
  num_threads = 16
  schema      = "SCHEMA"
  user        = "user"
  password    = "password"
}
