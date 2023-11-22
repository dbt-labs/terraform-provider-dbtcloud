resource "dbtcloud_environment_variable_job_override" "my_env_var_job_override" {
  name              = dbtcloud_environment_variable.dbt_my_env_var.name
  project_id        = dbtcloud_project.dbt_project.id
  job_definition_id = dbtcloud_job.daily_job.id
  raw_value         = "my_override_value"
}