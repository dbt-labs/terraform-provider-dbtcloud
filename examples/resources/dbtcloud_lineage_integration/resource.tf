// the resource can only be configured when a Prod environment has been set
// so, you might want to explicitly set the dependency on your Prod environment resource

resource "dbtcloud_lineage_integration" "my_lineage" {
  project_id = dbtcloud_project.my_project.id
  host       = "my.host.com"
  site_id    = "mysiteid"
  token_name = "my-token-name"
  token      = "my-sensitive-token"

  depends_on = [dbtcloud_environment.my_prod_env]
}
