// Using the classic sensitive attribute (stored in state)
resource "dbtcloud_databricks_platform_metadata_credential" "example" {
  connection_id = dbtcloud_global_connection.databricks.id

  catalog_ingestion_enabled = true
  cost_optimization_enabled = false
  cost_insights_enabled     = false

  token   = var.databricks_token
  catalog = "main"
}

// Using write-only attributes (not stored in state, requires Terraform >= 1.11)
variable "databricks_metadata_token" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_databricks_platform_metadata_credential" "example_wo" {
  connection_id = dbtcloud_global_connection.databricks.id

  catalog_ingestion_enabled = true
  cost_optimization_enabled = false
  cost_insights_enabled     = false

  token_wo         = var.databricks_metadata_token
  token_wo_version = 1
  catalog          = "main"
}

