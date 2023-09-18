// use dbt_cloud_repository instead of dbtcloud_repository for the legacy resource names
// legacy names will be removed from 0.3 onwards

# repo cloned via the GitHub integration, manually entering the `github_installation_id`
resource "dbtcloud_repository" "github_repo" {
  project_id             = dbtcloud_project.my_project.id
  remote_url             = "git@github.com:<github_org>/<github_repo>.git"
  github_installation_id = 9876
  git_clone_strategy     = "github_app"
}

# repo cloned via the GitHub integration, with auto-retrieval of the `github_installation_id`
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
  project_id             = dbtcloud_project.my_project.id
  remote_url             = "git@github.com:<github_org>/<github_repo>.git"
  github_installation_id = local.github_installation_id
  git_clone_strategy     = "github_app"
}


# repo cloned via the GitLab integration
# as of 15 Sept 2023 this resource requires using a user token and can't be set with a service token
resource "dbtcloud_repository" "gitlab_repo" {
  project_id        = dbtcloud_project.my_project_2.id
  remote_url        = "<gitlab-group>/<gitlab-project>"
  gitlab_project_id = 8765
}


# repo cloned via the deploy token strategy
resource "dbtcloud_repository" "deploy_repo" {
  project_id         = dbtcloud_project.my_project_3.id
  remote_url         = "git://github.com/<github_org>/<github_repo>.git"
  git_clone_strategy = "deploy_key"
}
