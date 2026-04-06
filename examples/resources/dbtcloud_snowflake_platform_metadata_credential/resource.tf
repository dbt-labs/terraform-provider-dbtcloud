// Using the classic sensitive attributes (stored in state)
resource "dbtcloud_snowflake_platform_metadata_credential" "password_auth" {
  connection_id = dbtcloud_global_connection.snowflake.id

  catalog_ingestion_enabled = true
  cost_optimization_enabled = true
  cost_insights_enabled     = false

  auth_type = "password"
  user      = "METADATA_READER"
  password  = var.snowflake_password
  role      = "METADATA_READER_ROLE"
  warehouse = "METADATA_WH"
}

resource "dbtcloud_snowflake_platform_metadata_credential" "keypair_auth" {
  connection_id = dbtcloud_global_connection.snowflake.id

  catalog_ingestion_enabled = true
  cost_optimization_enabled = false
  cost_insights_enabled     = false

  auth_type              = "keypair"
  user                   = "METADATA_READER"
  private_key            = var.snowflake_private_key
  private_key_passphrase = var.snowflake_private_key_passphrase
  role                   = "METADATA_READER_ROLE"
  warehouse              = "METADATA_WH"
}

// Using write-only attributes (not stored in state, requires Terraform >= 1.11)
variable "snowflake_metadata_password" {
  type      = string
  ephemeral = true
}

variable "snowflake_metadata_private_key" {
  type      = string
  ephemeral = true
}

variable "snowflake_metadata_private_key_passphrase" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_snowflake_platform_metadata_credential" "password_auth_wo" {
  connection_id = dbtcloud_global_connection.snowflake.id

  catalog_ingestion_enabled = true
  cost_optimization_enabled = true
  cost_insights_enabled     = false

  auth_type          = "password"
  user               = "METADATA_READER"
  password_wo        = var.snowflake_metadata_password
  password_wo_version = 1
  role               = "METADATA_READER_ROLE"
  warehouse          = "METADATA_WH"
}

resource "dbtcloud_snowflake_platform_metadata_credential" "keypair_auth_wo" {
  connection_id = dbtcloud_global_connection.snowflake.id

  catalog_ingestion_enabled = true
  cost_optimization_enabled = false
  cost_insights_enabled     = false

  auth_type                         = "keypair"
  user                              = "METADATA_READER"
  private_key_wo                    = var.snowflake_metadata_private_key
  private_key_wo_version            = 1
  private_key_passphrase_wo         = var.snowflake_metadata_private_key_passphrase
  private_key_passphrase_wo_version = 1
  role                              = "METADATA_READER_ROLE"
  warehouse                         = "METADATA_WH"
}

