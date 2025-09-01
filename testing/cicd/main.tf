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
  name                         = "Test Project SL"
  dbt_project_subdirectory     = "/dbt"
}

resource "dbtcloud_environment" "test_environment" {
  project_id  = dbtcloud_project.test_project.id
  name        = "Test Environment SL"
  dbt_version = "latest"
  type        = "deployment"
}

resource "dbtcloud_service_token" "test_service_token" {
  name = "Test Service Token SL"
}

resource "dbtcloud_snowflake_semantic_layer_credential" "test_credential" {
  configuration = {
    name            = "Test Credential SL"
    project_id      = dbtcloud_project.test_project.id
    adapter_version = "1.0"
  }
  credential = {
    auth_type = "password"
    database  = "test_db"
    warehouse = "test_wh"
    user      = "test_user"
    password  = "test_password"
    role      = "test_role"
    project_id = dbtcloud_project.test_project.id
    num_threads = 1
  }
}

resource "dbtcloud_semantic_layer_configuration" "test_semantic_layer_config" {
  project_id     = dbtcloud_project.test_project.id
  environment_id = dbtcloud_environment.test_environment.environment_id
}

resource "dbtcloud_semantic_layer_credential_service_token_mapping" "test_mapping" {
  service_token_id           = dbtcloud_service_token.test_service_token.id
  semantic_layer_credential_id = dbtcloud_snowflake_semantic_layer_credential.test_credential.id
}
