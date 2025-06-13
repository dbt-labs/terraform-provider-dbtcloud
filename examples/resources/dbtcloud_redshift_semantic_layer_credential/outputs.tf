output "sl_cred_redshift_credential_id" {
  description = "The ID of the Redshift Semantic Layer credential"
  value       = dbtcloud_redshift_semantic_layer_credential.test_redshift_semantic_layer_credential.id
} 