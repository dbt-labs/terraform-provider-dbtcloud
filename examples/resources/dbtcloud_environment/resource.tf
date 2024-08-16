resource "dbtcloud_environment" "ci_environment" {
  // the dbt_version is major.minor.0-latest , major.minor.0-pre or versionless (Beta on 15 Feb 2024, to always be on the latest dbt version)
  dbt_version   = "versionless"
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
  dbt_version = "versionless"
  name        = "Dev"
  project_id  = dbtcloud_project.dbt_project.id
  type        = "development"
  connection_id = dbtcloud_global_connection.my_other_global_connection
}
