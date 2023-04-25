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
    "git_provider_webhook" : false,
    "github_webhook" : false,
    "schedule" : false
  }
}
