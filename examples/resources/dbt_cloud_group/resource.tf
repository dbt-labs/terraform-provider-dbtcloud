resource "dbt_cloud_group" "test_group" {
    name = "Test Group"
    group_permissions {
        permission_set = "member"
        all_projects = true
    }
    group_permissions {
        permission_set = "developer"
        all_projects = false
        project_id = dbt_cloud_project.test_project.id
    }
}
