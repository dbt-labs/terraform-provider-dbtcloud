// use dbt_cloud_job instead of dbtcloud_job for the legacy resource names
// legacy names will be removed from 0.3 onwards

# a job that has github_webhook and git_provider_webhook 
# set to false will be categorized as a "Deploy Job"
resource "dbtcloud_job" "test" {
  environment_id = var.dbt_cloud_environment_id
  execute_steps = [
    "dbt build"
  ]
  generate_docs        = true
  is_active            = true
  name                 = "Daily job"
  num_threads          = 64
  project_id           = data.dbtcloud_project.test_project.id
  run_generate_sources = true
  target_name          = "default"
  triggers = {
    "custom_branch_only" : false,
    "github_webhook" : false,
    "git_provider_webhook" : false,
    "schedule" : true
  }
  # this is the default that gets set up when modifying jobs in the UI
  schedule_days = [0, 1, 2, 3, 4, 5, 6]
  schedule_type = "days_of_week"
  schedule_hours = [0]
}


# a job that has github_webhook and git_provider_webhook set 
# to true will be categorized as a "Continuous Integration Job"
resource "dbtcloud_job" "ci_job" {
  environment_id = var.my_ci_environment_id
  execute_steps = [
    "dbt build -s state:modified+ --fail-fast"
  ]
  generate_docs            = false
  deferring_environment_id = dbtcloud_environment.my_prod_env.environment_id
  name                     = "CI Job"
  num_threads              = 32
  project_id               = data.dbtcloud_project.test_project.id
  run_generate_sources     = false
  triggers = {
    "custom_branch_only" : true,
    "github_webhook" : true,
    "git_provider_webhook" : true,
    "schedule" : false
  }
  # this is the default that gets set up when modifying jobs in the UI
  # this is not going to be used when schedule is set to false
  schedule_days = [0, 1, 2, 3, 4, 5, 6]
  schedule_type = "days_of_week"
}