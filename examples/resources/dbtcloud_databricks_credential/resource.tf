// Using the classic sensitive attribute (stored in state)
resource "dbtcloud_databricks_credential" "my_databricks_cred" {
  project_id   = dbtcloud_project.dbt_project.id
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "databricks"
}

// Using write-only attributes (not stored in state, requires Terraform >= 1.11)
//
// The token_wo value is never persisted in the Terraform state file.
// Use token_wo_version to trigger an update when the token changes.
variable "databricks_token" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_databricks_credential" "my_databricks_cred_wo" {
  project_id       = dbtcloud_project.dbt_project.id
  token_wo         = var.databricks_token
  token_wo_version = 1
  schema           = "my_schema"
  adapter_type     = "databricks"
}
