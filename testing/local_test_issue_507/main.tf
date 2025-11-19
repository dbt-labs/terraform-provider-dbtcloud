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

variable "snowflake_account" {
  description = "Snowflake Account"
  type        = string
}

variable "snowflake_database" {
  description = "Snowflake Database"
  type        = string
  default     = "TEST_DATABASE"
}

variable "snowflake_warehouse" {
  description = "Snowflake Warehouse"
  type        = string
  default     = "TEST_WAREHOUSE"
}

variable "snowflake_oauth_client_id" {
  description = "Snowflake OAuth Client ID"
  type        = string
  sensitive   = true
}

variable "snowflake_oauth_client_secret" {
  description = "Snowflake OAuth Client Secret"
  type        = string
  sensitive   = true
}

# Create OAuth configuration for Snowflake
resource "dbtcloud_oauth_configuration" "test_oauth" {
  name               = "Test OAuth Configuration - Issue 507"
  oauth_provider     = "snowflake"
  client_id          = var.snowflake_oauth_client_id
  client_secret      = var.snowflake_oauth_client_secret
  account_identifier = var.snowflake_account
}

# Create global connection with OAuth configuration
resource "dbtcloud_global_connection" "test_snowflake_oauth" {
  name                   = "Test Snowflake OAuth Connection - Issue 507"
  oauth_configuration_id = dbtcloud_oauth_configuration.test_oauth.id
  snowflake = {
    account             = var.snowflake_account
    database            = var.snowflake_database
    warehouse           = var.snowflake_warehouse
    allow_sso           = true
    oauth_client_id     = var.snowflake_oauth_client_id
    oauth_client_secret = var.snowflake_oauth_client_secret
  }
}

# Create a test project to use the connection
resource "dbtcloud_project" "test_oauth_project" {
  name = "Test OAuth Project - Issue 507"
}

# Create an environment using the global connection
resource "dbtcloud_environment" "test_oauth_env" {
  project_id    = dbtcloud_project.test_oauth_project.id
  name          = "OAuth Test Environment"
  dbt_version   = "1.7"
  type          = "deployment"
  connection_id = dbtcloud_global_connection.test_snowflake_oauth.id
}

# Outputs to verify resource creation
output "oauth_configuration_id" {
  value = dbtcloud_oauth_configuration.test_oauth.id
}

output "global_connection_id" {
  value = dbtcloud_global_connection.test_snowflake_oauth.id
}

output "global_connection_adapter_version" {
  value       = dbtcloud_global_connection.test_snowflake_oauth.adapter_version
  description = "This is a read-only attribute that should not cause issues on subsequent applies"
}

output "project_id" {
  value = dbtcloud_project.test_oauth_project.id
}

output "environment_id" {
  value = dbtcloud_environment.test_oauth_env.environment_id
}

