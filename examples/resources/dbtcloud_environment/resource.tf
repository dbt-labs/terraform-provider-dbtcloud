resource "dbtcloud_environment" "ci_environment" {
  // the dbt_version is major.minor.0-latest , major.minor.0-pre or versionless (Beta on 15 Feb 2024, to always be on the latest dbt version)
  dbt_version   = "1.6.0-latest"
  name          = "CI"
  project_id    = dbtcloud_project.dbt_project.id
  type          = "deployment"
  credential_id = dbtcloud_snowflake_credential.ci_credential.credential_id
}

// we can also set a deployment environment as being the production one
resource "dbtcloud_environment" "prod_environment" {
  dbt_version     = "1.6.0-latest"
  name            = "Prod"
  project_id      = dbtcloud_project.dbt_project.id
  type            = "deployment"
  credential_id   = dbtcloud_snowflake_credential.prod_credential.credential_id
  deployment_type = "production"
}

// Creating a development environment
resource "dbtcloud_environment" "dev_environment" {
  dbt_version = "1.6.0-latest"
  name        = "Dev"
  project_id  = dbtcloud_project.dbt_project.id
  type        = "development"
}
