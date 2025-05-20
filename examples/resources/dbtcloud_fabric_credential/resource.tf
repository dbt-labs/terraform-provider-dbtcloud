# when using AD authentication
resource "dbtcloud_fabric_credential" "my_fabric_cred_ad" {
  project_id           = dbtcloud_project.dbt_project.id
  schema               = "my_schema"
  user                 = "my_user"
  password             = "my_password"
  schema_authorization = "abcd"
}

# when using service principal authentication
resource "dbtcloud_fabric_credential" "my_fabric_cred_serv_princ" {
  project_id           = dbtcloud_project.dbt_project.id
  schema               = "my_schema"
  client_id            = "my_client_id"
  tenant_id            = "my_tenant_id"
  client_secret        = "my_secret"
  schema_authorization = "abcd"
}