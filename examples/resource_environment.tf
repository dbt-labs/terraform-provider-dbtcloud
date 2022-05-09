resource "dbt_cloud_environment" "test_environment" {
  dbt_version   = "1.0.1"
  name          = "test"
  project_id    = data.dbt_cloud_project.test_project.project_id
  type          = "deployment"
  credential_id = dbt_cloud_snowflake_credential.new_credential.credential_id
}
