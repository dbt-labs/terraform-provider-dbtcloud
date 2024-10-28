resource "dbtcloud_snowflake_credential" "prod_credential" {
  project_id  = dbtcloud_project.dbt_project.id
  auth_type   = "password"
  num_threads = 16
  schema      = "SCHEMA"
  user        = "user"
  password    = "password"
}
