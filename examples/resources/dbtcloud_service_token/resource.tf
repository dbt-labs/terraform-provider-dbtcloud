// use dbt_cloud_service_token instead of dbtcloud_service_token for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_service_token" "test_service_token" {
    name = "Test Service Token"
    service_token_permissions {
        permission_set = "git_admin"
        all_projects = true
    }
    service_token_permissions {
        permission_set = "job_admin"
        all_projects = false
        project_id = dbt_cloud_project.test_project.id
    }
}

// permission_set accepts one of the following values:
// "account_admin","admin","database_admin","git_admin","team_admin","job_admin","job_viewer","analyst","developer","stakeholder","readonly","project_creator","account_viewer","metadata_only"
