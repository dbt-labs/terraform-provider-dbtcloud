resource "dbtcloud_postgres_semantic_layer_credential" "test_postgres_semantic_layer_credential" {
  configuration = {
    project_id = var.project_id
	name = "Postgres SL Credential"
	adapter_version = "postgres_v0"
  }
  credential = {
  	project_id = var.project_id
	username = var.username
    password = var.password
    semantic_layer_credential = true
  }
  
}