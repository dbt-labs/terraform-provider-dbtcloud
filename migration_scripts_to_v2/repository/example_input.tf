resource "dbtcloud_repository" "prod_repo" {
  project_id         = 1234
  remote_url         = "https://github.com/my-org/my-repo.git"
  git_clone_strategy = "github_app"
  fetch_deploy_key   = false
}

resource "dbtcloud_repository" "staging_repo" {
  project_id         = 2345
  remote_url         = "git@github.com:my-org/my-repo.git"
  git_clone_strategy = "deploy_key"
  fetch_deploy_key   = true
}

data "dbtcloud_repository" "existing_repo" {
  project_id       = 1234
  repository_id    = 5678
  fetch_deploy_key = false
}
