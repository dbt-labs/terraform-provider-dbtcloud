resource "dbtcloud_service_token" "test_service_token" {
  name = "Test Service Token"

  // Grant the service token `git_admin` permissions on all projects
  service_token_permissions {
    permission_set = "git_admin"
    all_projects   = true
  }

  // Grant the service token `job_admin` permissions on a specific project
  service_token_permissions {
    permission_set = "job_admin"
    all_projects   = false
    project_id     = dbtcloud_project.dbt_project.id
  }

  // Grant the service token `developer` permissions on all projects, 
  // but only in the `development` and `staging` environments
  //
  // NOTE: This is only configurable for certain `permission_set` values
  service_token_permissions {
    permission_set = "developer"
    all_projects   = true
    writable_environment_categories = [
      "development",
      "staging"
    ]
  }
}
