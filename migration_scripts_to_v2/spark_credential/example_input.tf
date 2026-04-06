resource "dbtcloud_spark_credential" "prod_credential" {
  project_id  = 1234
  target_name = "prod"
  token       = var.spark_token
  schema      = "my_schema"
}

resource "dbtcloud_spark_credential" "dev_credential" {
  project_id  = 1234
  target_name = "dev"
  token       = var.spark_token_dev
  schema      = "dev_schema"
}

data "dbtcloud_spark_credential" "existing" {
  project_id    = 1234
  credential_id = 5678
  target_name   = "prod"
}
