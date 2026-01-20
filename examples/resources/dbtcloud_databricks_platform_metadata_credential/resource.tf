# Example: Databricks Platform Metadata Credential
resource "dbtcloud_databricks_platform_metadata_credential" "example" {
  connection_id = dbtcloud_global_connection.databricks.id

  catalog_ingestion_enabled = true
  cost_optimization_enabled = false
  cost_insights_enabled     = false

  token   = var.databricks_token
  catalog = "main"
}

