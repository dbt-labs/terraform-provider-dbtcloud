output "sl_cred_service_token_mapping_id" {
  description = "The ID of the mapping between semantic layer credential and service token"
  value       = dbtcloud_semantic_layer_credential_service_token_mapping.test_mapping.id
} 