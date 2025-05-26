# when using sql authentication
resource "dbtcloud_synapse_credential" "my_synapse_cred_sql" {
  project_id           = dbtcloud_project.dbt_project.id
  authentication       = "sql"
  schema               = "my_schema"
  user                 = "my_user"
  password             = "my_password"
  schema_authorization = "abcd"
}

# when using AD authentication
resource "dbtcloud_synapse_credential" "my_synapse_cred_ad" {
  project_id           = dbtcloud_project.dbt_project.id
  authentication       = "ActiveDirectoryPassword"
  schema               = "my_schema"
  user                 = "my_user"
  password             = "my_password"
  schema_authorization = "abcd"
}

# when using service principal authentication
resource "dbtcloud_synapse_credential" "my_synapse_cred_serv_princ" {
  project_id           = dbtcloud_project.dbt_project.id
  authentication       = "ServicePrincipal"
  schema               = "my_schema"
  client_id            = "my_client_id"
  tenant_id            = "my_tenant_id"
  client_secret        = "my_secret"
  schema_authorization = "abcd"
}