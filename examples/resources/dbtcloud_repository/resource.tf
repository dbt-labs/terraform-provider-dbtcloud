// NOTE for customers using the LEGACY dbt_cloud provider:
// use dbt_cloud_repository instead of dbtcloud_repository for the legacy resource names
// legacy names will be removed from 0.3 onwards


### repo cloned via the GitHub integration, manually entering the `github_installation_id`
resource "dbtcloud_repository" "github_repo" {
  project_id             = dbtcloud_project.dbt_project.id
  remote_url             = "git@github.com:<github_org>/<github_repo>.git"
  github_installation_id = 9876
  git_clone_strategy     = "github_app"
}


### repo cloned via the GitHub integration, with auto-retrieval of the `github_installation_id`
# here, we assume that `token` and `host_url` are respectively accessible via `var.dbt_token` and `var.dbt_host_url`
# NOTE: the following requires connecting via a user token and can't be retrieved with a service token
data "http" "github_installations_response" {
  url = format("%s/v2/integrations/github/installations/", var.dbt_host_url)
  request_headers = {
    Authorization = format("Bearer %s", var.dbt_token)
  }
}

locals {
  github_installation_id = jsondecode(data.http.github_installations_response.response_body)[0].id
}

resource "dbtcloud_repository" "github_repo_other" {
  project_id             = dbtcloud_project.dbt_project.id
  remote_url             = "git@github.com:<github_org>/<github_repo>.git"
  github_installation_id = local.github_installation_id
  git_clone_strategy     = "github_app"
}


### repo cloned via the GitLab integration
# as of 15 Sept 2023 this resource requires using a user token and can't be set with a service token - CC-791
resource "dbtcloud_repository" "gitlab_repo" {
  project_id         = dbtcloud_project.dbt_project.id
  remote_url         = "<gitlab-group>/<gitlab-project>"
  gitlab_project_id  = 8765
  git_clone_strategy = "deploy_token"
}


### repo cloned via the deploy token strategy
resource "dbtcloud_repository" "deploy_repo" {
  project_id         = dbtcloud_project.dbt_project.id
  remote_url         = "git://github.com/<github_org>/<github_repo>.git"
  git_clone_strategy = "deploy_key"
}


### repo cloned via the Azure Dev Ops integration
resource "dbtcloud_repository" "ado_repo" {
  project_id = dbtcloud_project.dbt_project.id
  # the following values can be added manually (IDs can be retrieved from the ADO API) or via data sources
  # remote_url                              = "https://abc@dev.azure.com/abc/def/_git/my_repo"
  # azure_active_directory_project_id       = "12345678-1234-1234-1234-1234567890ab"
  # azure_active_directory_repository_id    = "87654321-4321-abcd-abcd-464327678642"
  remote_url                                = data.dbtcloud_azure_dev_ops_repository.my_devops_repo.remote_url
  azure_active_directory_repository_id      = data.dbtcloud_azure_dev_ops_repository.my_devops_repo.id
  azure_active_directory_project_id         = data.dbtcloud_azure_dev_ops_project.my_devops_project.id
  azure_bypass_webhook_registration_failure = false
  git_clone_strategy                        = "azure_active_directory_app"
}