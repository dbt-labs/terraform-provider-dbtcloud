// use dbt_cloud_environment instead of dbtcloud_environment for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_environment" "test_environment" {
  // the dbt_version is always major.minor.0-latest or major.minor.0-pre
  dbt_version   = "1.5.0-latest"
  name          = "test"
  project_id    = data.dbtcloud_project.test_project.id
  type          = "deployment"
  credential_id = dbt_cloud_snowflake_credential.new_credential.credential_id
}
