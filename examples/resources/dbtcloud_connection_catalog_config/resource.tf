# Example: Configure catalog filters for a Snowflake connection
resource "dbtcloud_connection_catalog_config" "snowflake_filters" {
  connection_id = dbtcloud_global_connection.snowflake.id

  # Only ingest from these databases
  database_allow = ["analytics", "reporting"]

  # Exclude staging schemas
  schema_deny = ["staging", "temp", "scratch"]

  # Exclude temporary tables
  table_deny = ["tmp_*", "temp_*"]

  # Exclude secret views
  view_deny = ["secret_*", "internal_*"]
}

# Example: Minimal configuration - just filter databases
resource "dbtcloud_connection_catalog_config" "minimal" {
  connection_id = dbtcloud_global_connection.snowflake.id

  database_allow = ["production"]
}

# Example: Full configuration with platform metadata credential
resource "dbtcloud_snowflake_platform_metadata_credential" "creds" {
  connection_id             = dbtcloud_global_connection.snowflake.id
  catalog_ingestion_enabled = true

  auth_type = "password"
  user      = var.snowflake_user
  password  = var.snowflake_password
  role      = var.snowflake_role
  warehouse = var.snowflake_warehouse
}

resource "dbtcloud_connection_catalog_config" "with_creds" {
  connection_id = dbtcloud_global_connection.snowflake.id

  database_allow = ["analytics", "reporting"]
  database_deny  = ["sandbox"]

  schema_allow = ["public", "dbt_*"]
  schema_deny  = ["information_schema", "pg_*"]

  table_deny = ["_tmp_*", "_staging_*"]
  view_deny  = ["_internal_*"]

  depends_on = [dbtcloud_snowflake_platform_metadata_credential.creds]
}

