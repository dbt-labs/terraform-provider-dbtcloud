// use dbt_cloud_snowflake_credential instead of dbtcloud_snowflake_credential for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_snowflake_credential" "new_credential" {
  project_id  = data.dbt_cloud_project.test_project.id
  auth_type   = "password"
  num_threads = 16
  schema      = "SCHEMA"
  user        = "user"
  password    = "password"
}
