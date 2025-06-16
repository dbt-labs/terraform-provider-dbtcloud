output "sl_cred_databricks_credential_id" {
  description = "The ID of the Databricks Semantic Layer credential"
  value       = dbtcloud_databricks_semantic_layer_credential.sl_cred_databricks_example.id
} 