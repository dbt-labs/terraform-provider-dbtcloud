# Case 1: top-level timeout_seconds only — will be converted to an execution block
resource "dbtcloud_job" "deploy_job" {
  environment_id = 1234
  execute_steps  = ["dbt build"]
  name           = "Daily Deploy"
  project_id     = 5678
  triggers = {
    "github_webhook"       : false
    "git_provider_webhook" : false
    "schedule"             : true
  }
  timeout_seconds = 3600
}

# Case 2: top-level timeout_seconds alongside an execution block that has no timeout
# — timeout_seconds will be moved into the execution block
resource "dbtcloud_job" "ci_job" {
  environment_id = 1234
  execute_steps  = ["dbt build -s state:modified+"]
  name           = "CI Job"
  project_id     = 5678
  triggers = {
    "github_webhook"       : true
    "git_provider_webhook" : true
    "schedule"             : false
  }
  timeout_seconds = 1800
  execution = {
    # timeout_seconds not yet set here
  }
}

# Case 3: both top-level timeout_seconds and execution.timeout_seconds present
# — top-level is removed; the execution block value is kept
resource "dbtcloud_job" "another_job" {
  environment_id = 1234
  execute_steps  = ["dbt run"]
  name           = "Another Job"
  project_id     = 5678
  triggers = {
    "schedule" : false
  }
  timeout_seconds = 7200
  execution = {
    timeout_seconds = 900
  }
}

# Singular data source: deferring_job_id will be removed
data "dbtcloud_job" "my_job" {
  job_id           = 1234
  deferring_job_id = 5678
}

# Plural data source: deferring_job_definition_id will be removed
data "dbtcloud_jobs" "all_jobs" {
  project_id                  = 5678
  deferring_job_definition_id = 1234
}
