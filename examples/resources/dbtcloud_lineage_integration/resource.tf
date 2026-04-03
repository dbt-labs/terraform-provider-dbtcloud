// the resource can only be configured when a Prod environment has been set
// so, you might want to explicitly set the dependency on your Prod environment resource

// Using the classic sensitive attribute (stored in state)
resource "dbtcloud_lineage_integration" "my_lineage" {
  project_id = dbtcloud_project.my_project.id
  host       = "my.host.com"
  site_id    = "mysiteid"
  token_name = "my-token-name"
  token      = "my-sensitive-token"

  depends_on = [dbtcloud_environment.my_prod_env]
}

// Using write-only attributes (not stored in state, requires Terraform >= 1.11)
//
// The token_wo value is never persisted in the Terraform state file.
// Use token_wo_version to trigger an update when the token changes.
variable "lineage_token" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_lineage_integration" "my_lineage_wo" {
  project_id       = dbtcloud_project.my_project.id
  host             = "my.host.com"
  site_id          = "mysiteid"
  token_name       = "my-token-name"
  token_wo         = var.lineage_token
  token_wo_version = 1

  depends_on = [dbtcloud_environment.my_prod_env]
}
