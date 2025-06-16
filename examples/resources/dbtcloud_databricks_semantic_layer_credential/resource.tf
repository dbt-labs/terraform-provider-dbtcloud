resource "dbtcloud_databricks_semantic_layer_credential" "sl_cred_databricks_example" {
  configuration = {
    project_id      = var.project_id
    name            = "Databricks SL Credential"
    adapter_version = "databricks_v0"
  }
  credential = {
    project_id                 = var.project_id
    catalog                    = var.catalog 
    token                      = var.token
    semantic_layer_credential  = true
  }
}