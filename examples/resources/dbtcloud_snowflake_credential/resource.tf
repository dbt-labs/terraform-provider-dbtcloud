// Using the classic sensitive attribute (stored in state)
resource "dbtcloud_snowflake_credential" "prod_credential" {
  project_id  = dbtcloud_project.dbt_project.id
  auth_type   = "password"
  num_threads = 16
  schema      = "SCHEMA"
  user        = "user"
  password    = "password"
}

// Using write-only attributes (not stored in state, requires Terraform >= 1.11)
//
// The password_wo value is never persisted in the Terraform state file.
// Use password_wo_version to trigger an update when the password changes.
variable "snowflake_password" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_snowflake_credential" "prod_credential_wo" {
  project_id          = dbtcloud_project.dbt_project.id
  auth_type           = "password"
  num_threads         = 16
  schema              = "SCHEMA"
  user                = "user"
  password_wo         = var.snowflake_password
  password_wo_version = 1
}
