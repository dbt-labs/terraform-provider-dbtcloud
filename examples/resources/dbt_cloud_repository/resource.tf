# In order to find the github_installation_id, you can log in to dbt Cloud, replace <dbt_cloud_url> by your dbt Cloud URL and run the following commands in the Google Chrome console:

# ```
# dbt_cloud_api_result = await (fetch('https://<dbt_cloud_url>/api/v2/integrations/github/installations/').then(res => res.json()));
# console.log("github_application_id: " + dbt_cloud_api_result.filter(res => res["access_tokens_url"].includes("github"))[0]["id"]);
# ```

# Alternatively, you can go to the page https://<dbt_cloud_url>/api/v2/integrations/github/installations/ and read the value of "id"

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