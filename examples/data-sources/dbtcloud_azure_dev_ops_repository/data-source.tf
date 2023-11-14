data "dbtcloud_azure_dev_ops_repository" "my_ado_repository" {
  name = "my-repo-name"
  azure_dev_ops_project_id = data.dbtcloud_azure_dev_ops_project.my_ado_project.id
}