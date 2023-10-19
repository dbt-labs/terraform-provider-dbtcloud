// use dbt_cloud_group instead of dbtcloud_group for the legacy resource names
// legacy names will be removed from 0.3 onwards

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
