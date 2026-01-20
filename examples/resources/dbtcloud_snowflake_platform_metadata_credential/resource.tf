# Example: Snowflake Platform Metadata Credential with password auth
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

# Example: Snowflake Platform Metadata Credential with keypair auth
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

