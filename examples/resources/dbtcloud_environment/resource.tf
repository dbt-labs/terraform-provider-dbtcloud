// NOTE for customers using the LEGACY dbt_cloud provider:
// use dbt_cloud_environment instead of dbtcloud_environment for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_environment" "ci_environment" {
  // the dbt_version is always major.minor.0-latest or major.minor.0-pre
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
