resource "dbtcloud_postgres_credential" "postgres_prod_credential" {
  is_active      = true
  project_id     = dbtcloud_project.dbt_project.id
  type           = "postgres"
  default_schema = "my_schema"
  username       = "my_username"
  password       = "my_password"
  num_threads    = 16
}