// use dbt_cloud_environment_variable instead of dbtcloud_environment_variable for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_environment_variable" "my_env_var" {
  name       = "DBT_MY_ENV_VAR"
  project_id = dbtcloud_project.my_project.id
  environment_values = {
    "project" : "my_project_level_value",
    "My Env" : "my_env_level_value"
  }
  depends_on = [
    dbtcloud_project.my_project,
    dbtcloud_environment.my_env
  ]
}