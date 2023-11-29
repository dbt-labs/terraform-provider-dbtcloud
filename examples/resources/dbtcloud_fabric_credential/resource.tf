# when using AD authentication
resource "dbtcloud_databricks_credential" "my_databricks_cred" {
  project_id           = dbtcloud_project.dbt_project.id
  adapter_id           = dbtcloud_fabric_connection.my_fabric_connection.adapter_id
  schema               = "my_schema"
  user                 = "my_user"
  password             = "my_password"
  schema_authorization = "abcd"
}

# when using service principal authentication
resource "dbtcloud_databricks_credential" "my_spark_cred" {
  project_id           = dbtcloud_project.dbt_project.id
  adapter_id           = dbtcloud_fabric_connection.my_fabric_connection.adapter_id
  schema               = "my_schema"
  client_id            = "my_client_id"
  tenant_id            = "my_tenant_id"
  client_secret        = "my_secret"
  schema_authorization = "abcd"
}