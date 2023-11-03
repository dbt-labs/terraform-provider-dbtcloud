// NOTE for customers using the LEGACY dbt_cloud provider:
// use dbt_cloud_postgres_credential instead of dbtcloud_postgres_credential for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_postgres_credential" "postgres_prod_credential" {
  is_active      = true
  project_id     = dbtcloud_project.dbt_project.id
  type           = "postgres"
  default_schema = "my_schema"
  username       = "my_username"
  password       = "my_password"
  num_threads    = 16
}