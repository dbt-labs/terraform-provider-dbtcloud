resource "dbt_cloud_job" "test" {
  environment_id = var.dbt_cloud_environment_id
  execute_steps = [
    "dbt test"
  ]
  generate_docs        = false
  is_active            = true
  name                 = "Test"
  num_threads          = 64
  project_id           = data.dbt_cloud_project.test_project.id
  run_generate_sources = false
  target_name          = "default"
  triggers = {
    "custom_branch_only" : true,
    "github_webhook" : false,
    "git_provider_webhook" : false,
    "schedule" : false
  }
  # this is the default that gets set up when modifying jobs in the UI
  schedule_days = [0,1,2,3,4,5,6]
  schedule_type = "days_of_week"
}
