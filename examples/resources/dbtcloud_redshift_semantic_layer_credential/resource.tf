resource "dbtcloud_redshift_semantic_layer_credential" "test_redshift_semantic_layer_credential" {
  configuration = {
    project_id = var.project_id
	name = "Redshift SL Credential"
	adapter_version = "redshift_v0"
  }
  credential = {
  	project_id = var.project_id
	username = var.username
	is_active = true
	password = var.password
	num_threads = var.num_threads
	default_schema = var.default_schema
  }
  
}