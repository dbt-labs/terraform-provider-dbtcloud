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
