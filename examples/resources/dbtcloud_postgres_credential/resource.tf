// Using the classic sensitive attribute (stored in state)
resource "dbtcloud_postgres_credential" "postgres_prod_credential" {
  is_active      = true
  project_id     = dbtcloud_project.dbt_project.id
  type           = "postgres"
  default_schema = "my_schema"
  username       = "my_username"
  password       = "my_password"
  num_threads    = 16
}

// Using write-only attributes (not stored in state, requires Terraform >= 1.11)
//
// The password_wo value is never persisted in the Terraform state file.
// Use password_wo_version to trigger an update when the password changes.
variable "postgres_password" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_postgres_credential" "postgres_prod_credential_wo" {
  is_active           = true
  project_id          = dbtcloud_project.dbt_project.id
  type                = "postgres"
  default_schema      = "my_schema"
  username            = "my_username"
  password_wo         = var.postgres_password
  password_wo_version = 1
  num_threads         = 16
}
