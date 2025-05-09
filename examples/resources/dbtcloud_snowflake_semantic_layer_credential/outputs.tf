output "password_auth_credential_id" {
  description = "The ID of the Snowflake Semantic Layer credential with password auth"
  value       = dbtcloud_snowflake_semantic_layer_credential.password_auth.id
}

output "keypair_auth_credential_id" {
  description = "The ID of the Snowflake Semantic Layer credential with key pair auth"
  value       = dbtcloud_snowflake_semantic_layer_credential.keypair_auth.id
} 