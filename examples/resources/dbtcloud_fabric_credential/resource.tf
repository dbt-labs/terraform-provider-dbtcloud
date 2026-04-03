# Using the classic sensitive attributes (stored in state)

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

// Using write-only attributes (not stored in state, requires Terraform >= 1.11)
//
// The password_wo and client_secret_wo values are never persisted in the Terraform state file.
// Use password_wo_version / client_secret_wo_version to trigger an update when the secret changes.
variable "fabric_password" {
  type      = string
  ephemeral = true
}

variable "fabric_client_secret" {
  type      = string
  ephemeral = true
}

# when using AD authentication with write-only password
resource "dbtcloud_fabric_credential" "my_fabric_cred_ad_wo" {
  project_id           = dbtcloud_project.dbt_project.id
  schema               = "my_schema"
  user                 = "my_user"
  password_wo          = var.fabric_password
  password_wo_version  = 1
  schema_authorization = "abcd"
}

# when using service principal authentication with write-only client secret
resource "dbtcloud_fabric_credential" "my_fabric_cred_serv_princ_wo" {
  project_id               = dbtcloud_project.dbt_project.id
  schema                   = "my_schema"
  client_id                = "my_client_id"
  tenant_id                = "my_tenant_id"
  client_secret_wo         = var.fabric_client_secret
  client_secret_wo_version = 1
  schema_authorization     = "abcd"
}
