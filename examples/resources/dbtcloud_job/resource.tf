# a job that has github_webhook and git_provider_webhook 
# set to false will be categorized as a "Deploy Job"
resource "dbtcloud_job" "daily_job" {
  environment_id = dbtcloud_environment.prod_environment.environment_id
  execute_steps = [
    "dbt build"
  ]
  generate_docs        = true
  is_active            = true
  name                 = "Daily job"
  num_threads          = 64
  project_id           = dbtcloud_project.dbt_project.id
  run_generate_sources = true
  target_name          = "default"
  triggers = {
    "github_webhook" : false
    "git_provider_webhook" : false
    "schedule" : true
    "on_merge" : false
  }
  # this is the default that gets set up when modifying jobs in the UI
  schedule_days  = [0, 1, 2, 3, 4, 5, 6]
  schedule_type  = "days_of_week"
  schedule_hours = [0]
}


# a job that has github_webhook and git_provider_webhook set 
# to true will be categorized as a "Continuous Integration Job"
resource "dbtcloud_job" "ci_job" {
  environment_id = dbtcloud_environment.ci_environment.environment_id
  execute_steps = [
    "dbt build -s state:modified+ --fail-fast"
  ]
  generate_docs            = false
  deferring_environment_id = dbtcloud_environment.prod_environment.environment_id
  name                     = "CI Job"
  num_threads              = 32
  project_id               = dbtcloud_project.dbt_project.id
  run_generate_sources     = false
  triggers = {
    "github_webhook" : true
    "git_provider_webhook" : true
    "schedule" : false
    "on_merge" : false
  }
  # this is the default that gets set up when modifying jobs in the UI
  # this is not going to be used when schedule is set to false
  schedule_days = [0, 1, 2, 3, 4, 5, 6]
  schedule_type = "days_of_week"
}

# a job that is set to be triggered after another job finishes
# this is sometimes referred as 'job chaining'
resource "dbtcloud_job" "downstream_job" {
  environment_id = dbtcloud_environment.project2_prod_environment.environment_id
  execute_steps = [
    "dbt build -s +my_model"
  ]
  generate_docs            = true
  name                     = "Downstream job in project 2"
  num_threads              = 32
  project_id               = dbtcloud_project.dbt_project2.id
  run_generate_sources     = true
  triggers = {
    "github_webhook" : false
    "git_provider_webhook" : false
    "schedule" : false
    "on_merge" : false
  }
  schedule_days = [0, 1, 2, 3, 4, 5, 6]
  schedule_type = "days_of_week"
  job_completion_trigger_condition {
    job_id = dbtcloud_job.daily_job.id
    project_id = dbtcloud_project.dbt_project.id
    statuses = ["success"]
  }
}
