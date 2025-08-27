terraform {
  required_providers {
    dbtcloud = {
      source  = "dbt-labs/dbtcloud"
      version = ">= 0.2.0"
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
  dbt_version = "1.7.0"
  type       = "development"
}

resource "dbtcloud_repository" "test_repository" {
  project_id = dbtcloud_project.test_project.id
  remote_url = "https://github.com/dbt-labs/dbt-starter-project"
  git_clone_strategy = "github_app"
}

resource "dbtcloud_job" "test_job" {
  project_id = dbtcloud_project.test_project.id
  environment_id = dbtcloud_environment.test_environment.id
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
