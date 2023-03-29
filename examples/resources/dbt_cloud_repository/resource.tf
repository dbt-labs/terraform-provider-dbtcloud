resource "dbt_cloud_repository" "github_repo" {
  project_id = 1234
  remote_url = "git://github.com/<github_org>/<github_repo>.git"
  github_installation_id = 9876
  git_clone_strategy = "github_app"
}


resource "dbt_cloud_repository" "gitlab_repo" {
  project_id = 2345
  remote_url = "<gitlab-group>/<gitlab-project>"
  gitlab_project_id = 8765
}


resource "dbt_cloud_repository" "deploy_repo" {
  project_id = 3456
  remote_url = "git://github.com/<github_org>/<github_repo>.git"
  git_clone_strategy = "deploy_token"
}
