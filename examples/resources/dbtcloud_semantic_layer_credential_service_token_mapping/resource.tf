resource "dbtcloud_semantic_layer_credential_service_token_mapping" "test_mapping" {
  semantic_layer_credential_id = dbtcloud_redshift_semantic_layer_credential.test.id
  service_token_id = dbtcloud_service_token.test_service_token.id
  project_id = dbtcloud_project.test_project.id
}