// Secrets are write-only (not stored in state, requires Terraform >= 1.11).
// Change the matching `_wo_version` to send a new value.
variable "snowflake_private_key" {
  type      = string
  ephemeral = true
}

variable "snowflake_private_key_passphrase" {
  type      = string
  ephemeral = true
}

variable "databricks_token" {
  type      = string
  ephemeral = true
}

// The adapter block must match the adapter of the connection
resource "dbtcloud_analytics_credential" "snowflake" {
  connection_id = dbtcloud_global_connection.snowflake.id

  snowflake = {
    auth_type                         = "keypair"
    user                              = "ANALYTICS_USER"
    role                              = "ANALYTICS_ROLE"
    warehouse                         = "ANALYTICS_WH"
    private_key_wo                    = var.snowflake_private_key
    private_key_wo_version            = 1
    private_key_passphrase_wo         = var.snowflake_private_key_passphrase
    private_key_passphrase_wo_version = 1
  }
}

resource "dbtcloud_analytics_credential" "databricks" {
  connection_id = dbtcloud_global_connection.databricks.id

  databricks = {
    token_wo         = var.databricks_token
    token_wo_version = 1
    catalog          = "analytics"
  }
}
