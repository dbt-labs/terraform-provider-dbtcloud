# repo cloned via the GitHub integration, manually entering the `github_installation_id`
resource "dbt_cloud_repository" "github_repo" {
  project_id = 1234
  remote_url = "git@github.com:<github_org>/<github_repo>.git"
  github_installation_id = 9876
  git_clone_strategy = "github_app"
}

# repo cloned via the GitHub integration, with auto-retrieval of the `github_installation_id`
# here, we assume that `token` and `host_url` are respectively accessible via `var.dbt_token` and `var.dbt_host_url`
data "http" "github_installations_reponse" {
  url = format("%s/v2/integrations/github/installations/", var.dbt_host_url)
  request_headers = {
    Authorization = format("Bearer %s", var.dbt_token)
  }
}

locals {
  github_installation_id = jsondecode(data.http.github_installations_reponse.response_body)[0].id
}

resource "dbt_cloud_repository" "github_repo_other" {
  project_id = 1234
  remote_url = "git@github.com:<github_org>/<github_repo>.git"
  github_installation_id = local.github_installation_id
  git_clone_strategy = "github_app"
}


# repo cloned via the GitLab integration
resource "dbt_cloud_repository" "gitlab_repo" {
  project_id = 2345
  remote_url = "<gitlab-group>/<gitlab-project>"
  gitlab_project_id = 8765
}


# repo cloned via the deploy token strategy
resource "dbt_cloud_repository" "deploy_repo" {
  project_id = 3456
  remote_url = "git://github.com/<github_org>/<github_repo>.git"
  git_clone_strategy = "deploy_token"
}
