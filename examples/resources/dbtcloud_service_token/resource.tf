// NOTE for customers using the LEGACY dbt_cloud provider:
// use dbt_cloud_service_token instead of dbtcloud_service_token for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_service_token" "test_service_token" {
  name = "Test Service Token"
  service_token_permissions {
    permission_set = "git_admin"
    all_projects   = true
  }
  service_token_permissions {
    permission_set = "job_admin"
    all_projects   = false
    project_id     = dbtcloud_project.dbt_project.id
  }
}
