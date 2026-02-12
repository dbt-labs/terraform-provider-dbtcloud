terraform {
  required_providers {
    dbtcloud = {
      source = "dbt-labs/dbtcloud"
    }
  }
}

provider "dbtcloud" {}

# -----------------------------------------------------------------------------
# Project
# -----------------------------------------------------------------------------
resource "dbtcloud_project" "eric_tf_profile_test" {
  name = "ERIC-TF-PROFILE-TEST"
}

# -----------------------------------------------------------------------------
# Global Connection (Snowflake)
# -----------------------------------------------------------------------------
resource "dbtcloud_global_connection" "eric_tf_profile_test_conn" {
  name = "ERIC-TF-PROFILE-TEST-CONN"

  snowflake = {
    account   = "test-account"
    warehouse = "test-warehouse"
    database  = "test-database"
  }
}

# -----------------------------------------------------------------------------
# Credential (Snowflake)
# -----------------------------------------------------------------------------
resource "dbtcloud_snowflake_credential" "eric_tf_profile_test_cred" {
  is_active   = true
  project_id  = dbtcloud_project.eric_tf_profile_test.id
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
# Profile — ties connection + credential together
# -----------------------------------------------------------------------------
resource "dbtcloud_profile" "eric_tf_profile_test" {
  project_id     = dbtcloud_project.eric_tf_profile_test.id
  key            = "eric-tf-profile-test-key"
  connection_id  = dbtcloud_global_connection.eric_tf_profile_test_conn.id
  credentials_id = dbtcloud_snowflake_credential.eric_tf_profile_test_cred.credential_id
}

# -----------------------------------------------------------------------------
# Deployment Environment — bound to the profile via primary_profile_id
# -----------------------------------------------------------------------------
resource "dbtcloud_environment" "eric_tf_profile_test_deploy" {
  name               = "ERIC-TF-PROFILE-TEST-DEPLOY"
  type               = "deployment"
  dbt_version        = "latest"
  project_id         = dbtcloud_project.eric_tf_profile_test.id
  deployment_type    = "production"
  primary_profile_id = dbtcloud_profile.eric_tf_profile_test.profile_id
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------
output "project_id" {
  value = dbtcloud_project.eric_tf_profile_test.id
}

output "profile_id" {
  value = dbtcloud_profile.eric_tf_profile_test.profile_id
}

output "environment_id" {
  value = dbtcloud_environment.eric_tf_profile_test_deploy.environment_id
}
