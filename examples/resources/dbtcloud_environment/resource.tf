resource "dbtcloud_environment" "ci_environment" {
  // the dbt_version is major.minor.0-latest , major.minor.0-pre, compatible, extended, versionless, latest or latest-fusion (by default, it is set to latest if not configured)
  dbt_version   = "latest"
  name          = "CI"
  project_id    = dbtcloud_project.dbt_project.id
  type          = "deployment"
  credential_id = dbtcloud_snowflake_credential.ci_credential.credential_id
  connection_id = dbtcloud_global_connection.my_global_connection.id
}

// we can also set a deployment environment as being the production one
resource "dbtcloud_environment" "prod_environment" {
  dbt_version     = "1.7.0-latest"
  name            = "Prod"
  project_id      = dbtcloud_project.dbt_project.id
  type            = "deployment"
  credential_id   = dbtcloud_snowflake_credential.prod_credential.credential_id
  deployment_type = "production"
  connection_id   = dbtcloud_connection.my_legacy_connection.connection_id
}

// Creating a development environment
resource "dbtcloud_environment" "dev_environment" {
  dbt_version = "latest"
  name        = "Dev"
  project_id  = dbtcloud_project.dbt_project.id
  type        = "development"
  connection_id = dbtcloud_global_connection.my_other_global_connection.id
  // credential_id is not actionable for development environments
}

// Deployment environment with a primary profile (binds connection + credentials via profile)
// NOTE: avoid setting connection_id, credential_id, or extended_attributes_id alongside
// primary_profile_id — dbt Cloud may propagate the environment's values onto the profile,
// overwriting the profile's own settings and affecting other environments sharing that profile.
resource "dbtcloud_environment" "profiled_environment" {
  dbt_version        = "latest"
  name               = "Staging"
  project_id         = dbtcloud_project.dbt_project.id
  type               = "deployment"
  deployment_type    = "staging"
  primary_profile_id = dbtcloud_profile.my_profile.profile_id
}
