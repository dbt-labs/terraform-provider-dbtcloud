# Using the classic sensitive attribute (stored in state)
resource "dbtcloud_clickhouse_credential" "my_clickhouse_cred" {
  project_id  = dbtcloud_project.dbt_project.id
  schema      = "my_schema"
  user        = "my_user"
  password    = "my_password"
  target_name = "default"
  threads     = 4
}

// Using the write-only attribute (not stored in state, requires Terraform >= 1.11)
//
// The password_wo value is never persisted in the Terraform state file.
// Use password_wo_version to trigger an update when the password changes.
variable "clickhouse_password" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_clickhouse_credential" "my_clickhouse_cred_wo" {
  project_id          = dbtcloud_project.dbt_project.id
  schema              = "my_schema"
  user                = "my_user"
  password_wo         = var.clickhouse_password
  password_wo_version = 1
  target_name         = "default"
  threads             = 4
}
