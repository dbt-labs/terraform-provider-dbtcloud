terraform {
  required_providers {
    dbtcloud = {
      source  = "dbt-labs/dbtcloud"
      version = ">= 1.2.0"
    }
  }
}

provider "dbtcloud" {
  account_id = var.dbt_cloud_account_id
  token      = var.dbt_cloud_token
  host_url   = var.dbt_cloud_host_url
}

variable "dbt_cloud_account_id" {
  description = "dbt Cloud Account ID"
  type        = number
}

variable "dbt_cloud_token" {
  description = "dbt Cloud Token"
  type        = string
  sensitive   = true
}

variable "dbt_cloud_host_url" {
  description = "dbt Cloud Host URL"
  type        = string
  default     = "https://cloud.getdbt.com/api"
}

resource "dbtcloud_project" "test_project" {
  name        = "Test Project"
  dbt_project_subdirectory = "/dbt"
}

resource "dbtcloud_environment" "test_environment" {
  project_id = dbtcloud_project.test_project.id
  name       = "Test Environment"
  dbt_version = "latest-fusion"
  type       = "development"
}

resource "dbtcloud_repository" "test_repository" {
  project_id = dbtcloud_project.test_project.id
  remote_url = "git@github.com:dbt-labs/tf-provider-e2e-test.git"
  git_clone_strategy = "deploy_key"
}

resource "dbtcloud_job" "test_job" {
  project_id = dbtcloud_project.test_project.id
  environment_id = dbtcloud_environment.test_environment.environment_id
  name = "Test Job"
  execute_steps = [
    "dbt run"
  ]
  triggers = {
    github_webhook = false
    git_provider_webhook = false
    schedule = false
  }
}

resource "dbtcloud_service_token" "test_service_token" {
  name = "Test Service Token"
}

resource "dbtcloud_group" "test_group" {
  name = "Test Group"
}

resource "dbtcloud_environment_variable" "test_env_var" {
  project_id = dbtcloud_project.test_project.id
  name       = "DBT_TEST_ENV_VAR"
  environment_values = {
    "project"                           = "default_value"
    dbtcloud_environment.test_environment.name = "test_value"
  }
  depends_on = [
    dbtcloud_environment.test_environment
  ]
}

resource "dbtcloud_webhook" "test_webhook" {
  name        = "Test Webhook"
  description = "A webhook for testing"
  client_url  = "https://example.com/webhook"
  event_types = [
    "job.run.started",
    "job.run.completed"
  ]
}

resource "dbtcloud_notification" "test_notification_success" {
  user_id    = 100 # Using a placeholder user ID
  on_success = [dbtcloud_job.test_job.id]
  state      = 1 # active
}

resource "dbtcloud_notification" "test_notification_failure" {
  user_id    = 100 # Using a placeholder user ID
  on_failure = [dbtcloud_job.test_job.id]
  state      = 1 # active
  notification_type = 4 # PagerDuty
}
