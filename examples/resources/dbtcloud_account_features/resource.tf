# Manage account-level feature flags in dbt Cloud
resource "dbtcloud_account_features" "my_features" {
  # CI/CD features
  advanced_ci     = true
  partial_parsing = true
  repo_caching    = true

  # AI and insights features
  ai_features   = true
  cost_insights = true

  # Catalog/Explorer features
  catalog_ingestion   = true
  explorer_account_ui = true

  # Migration features
  fusion_migration_permissions = false

  # Warehouse features
  warehouse_cost_visibility = true
}

