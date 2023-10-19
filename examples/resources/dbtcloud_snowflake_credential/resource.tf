// use dbt_cloud_snowflake_credential instead of dbtcloud_snowflake_credential for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_snowflake_credential" "prod_credential" {
  project_id  = data.dbtcloud_project.dbt_project.id
  auth_type   = "password"
  num_threads = 16
  schema      = "SCHEMA"
  user        = "user"
  password    = "password"
}
