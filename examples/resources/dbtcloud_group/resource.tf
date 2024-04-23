resource "dbtcloud_group" "tf_group_1" {
  name = "TF Group 1"
  group_permissions {
    permission_set = "member"
    all_projects   = true
  }
  group_permissions {
    permission_set = "developer"
    all_projects   = false
    project_id     = dbtcloud_project.dbt_project.id
  }
}
