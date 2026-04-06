resource "dbtcloud_databricks_credential" "prod_credential" {
  project_id    = 1234
  adapter_type  = "databricks"
  target_name   = "prod"
  token         = var.databricks_token
  schema        = "my_schema"
  catalog       = "my_catalog"
}

resource "dbtcloud_databricks_credential" "dev_credential" {
  project_id    = 1234
  adapter_type  = "spark"
  target_name   = "dev"
  token         = var.databricks_token_dev
  schema        = "dev_schema"
}

data "dbtcloud_databricks_credential" "existing" {
  project_id    = 1234
  credential_id = 5678
  target_name   = "prod"
}
