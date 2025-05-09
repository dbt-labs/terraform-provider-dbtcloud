# Example of Snowflake Semantic Layer Credential with password authentication
resource "dbtcloud_snowflake_semantic_layer_credential" "password_auth" {
  configuration = {
    project_id      = var.project_id
    name            = "Snowflake SL Credential - Password Auth"
    adapter_version = "snowflake_v0"
  }
  credential = {
    project_id                 = var.project_id
    is_active                  = true
    auth_type                  = "password"
    database                   = var.database
    schema                     = var.schema
    warehouse                  = var.warehouse
    role                       = var.role
    user                       = var.user
    password                   = var.password
    num_threads                = 4
    semantic_layer_credential  = true
  }
}

# Example of Snowflake Semantic Layer Credential with key pair authentication
resource "dbtcloud_snowflake_semantic_layer_credential" "keypair_auth" {
  configuration = {
    project_id      = var.project_id
    name            = "Snowflake SL Credential - Key Pair Auth"
    adapter_version = "snowflake_v0"
  }
  credential = {
    project_id                 = var.project_id
    is_active                  = true
    auth_type                  = "keypair"
    database                   = var.database
    schema                     = var.schema
    warehouse                  = var.warehouse
    role                       = var.role
    private_key                = var.private_key
    private_key_passphrase     = var.private_key_passphrase
    num_threads                = 4
    semantic_layer_credential  = true
  }
} 