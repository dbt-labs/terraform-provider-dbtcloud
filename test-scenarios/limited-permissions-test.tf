# Test Configuration for Limited Permissions Scenario
# This tests the fix for issue #537

terraform {
  required_providers {
    dbtcloud = {
      source = "dbt-labs/dbtcloud"
      # When using dev_overrides, version is ignored
      version = "~> 1.0"
    }
  }
}

provider "dbtcloud" {
  # Set these via environment variables:
  # export DBT_CLOUD_ACCOUNT_ID=<your_account_id>
  # export DBT_CLOUD_TOKEN=<your_limited_token>
  # export DBT_CLOUD_HOST_URL=<your_host_url>  # optional
}

# Replace these values with your actual IDs
variable "project_id" {
  description = "ID of your dbt Cloud project"
  type        = number
  # default     = 12345  # Uncomment and set your project ID
}

variable "environment_with_access" {
  description = "Environment where your token HAS write access"
  type        = number
  # default     = 54321  # Uncomment and set your environment ID
}

variable "environment_without_access" {
  description = "Environment where your token LACKS write access"
  type        = number
  # default     = 98765  # Uncomment and set your environment ID
}

# Create a job in the environment where you have access
resource "dbtcloud_job" "test_job" {
  name           = "Test Job - Permission Fix"
  project_id     = var.project_id
  environment_id = var.environment_with_access  # Change to environment_without_access to test the error
  
  execute_steps = [
    "dbt run"
  ]
  
  triggers {
    schedule = false
    github_webhook = false
    git_provider_webhook = false
    on_merge = false
  }
  
  is_active = true
}

output "job_id" {
  value       = dbtcloud_job.test_job.job_id
  description = "The ID of the created job"
}

output "environment_id" {
  value       = dbtcloud_job.test_job.environment_id
  description = "The environment ID of the job"
}

