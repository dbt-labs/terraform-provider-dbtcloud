output "sl_cred_postgres_credential_id" {
  description = "The ID of the Postgres Semantic Layer credential"
  value       = dbtcloud_postgres_semantic_layer_credential.test_postgres_semantic_layer_credential.id
} 