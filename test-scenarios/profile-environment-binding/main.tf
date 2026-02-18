terraform {
  required_providers {
    dbtcloud = {
      source = "dbt-labs/dbtcloud"
    }
  }
}

provider "dbtcloud" {}

# =============================================================================
# NOTE: This example uses placeholder credentials for demonstration purposes.
# Never hardcode real Snowflake credentials (or any secrets) in Terraform files.
# Instead, use environment variables (TF_VAR_*) or a terraform.tfvars file that
# is excluded from version control via .gitignore.
#
# This configuration creates a minimal project with a profile-based environment
# binding. It does NOT set up a repository connection or a development
# environment, so the resulting project will not be fully functional for
# running dbt jobs without additional configuration.
# =============================================================================

# -----------------------------------------------------------------------------
# Project
# -----------------------------------------------------------------------------
resource "dbtcloud_project" "example" {
  name = "Snowflake Profile Example"
}

# -----------------------------------------------------------------------------
# Global Connection (Snowflake)
# -----------------------------------------------------------------------------
resource "dbtcloud_global_connection" "example" {
  name = "Snowflake Example Connection"

  snowflake = {
    account   = "test-account"
    warehouse = "test-warehouse"
    database  = "test-database"
  }
}

# -----------------------------------------------------------------------------
# Credential (Snowflake)
# -----------------------------------------------------------------------------
resource "dbtcloud_snowflake_credential" "example" {
  is_active   = true
  project_id  = dbtcloud_project.example.id
  auth_type   = "password"
  database    = "test-database"
  role        = "test-role"
  warehouse   = "test-warehouse"
  schema      = "test_schema"
  user        = "test-user"
  password    = "test-password"
  num_threads = 3
}

# -----------------------------------------------------------------------------
# Profile — ties a connection and credential together under a key.
# The key acts as a logical name for the warehouse target (e.g. "snowflake_prod").
# -----------------------------------------------------------------------------
resource "dbtcloud_profile" "example" {
  project_id     = dbtcloud_project.example.id
  key            = "snowflake_example"
  connection_id  = dbtcloud_global_connection.example.id
  credentials_id = dbtcloud_snowflake_credential.example.credential_id
}

# -----------------------------------------------------------------------------
# Deployment Environment — bound to the profile via primary_profile_id
# -----------------------------------------------------------------------------
resource "dbtcloud_environment" "production" {
  name               = "Production"
  type               = "deployment"
  dbt_version        = "latest"
  project_id         = dbtcloud_project.example.id
  deployment_type    = "production"
  primary_profile_id = dbtcloud_profile.example.profile_id
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------
output "project_id" {
  value = dbtcloud_project.example.id
}

output "profile_id" {
  value = dbtcloud_profile.example.profile_id
}

output "environment_id" {
  value = dbtcloud_environment.production.environment_id
}
