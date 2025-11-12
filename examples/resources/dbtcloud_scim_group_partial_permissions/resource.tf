# Retrieve the SCIM-managed group
data "dbtcloud_groups" "engineering" {
  name = "Engineering Team"  # This group is synced from your IdP via SCIM
}

# Get the project to apply permissions to
data "dbtcloud_project" "my_project" {
  name = "My Analytics Project"
}

# Platform team can manage base account permissions
resource "dbtcloud_scim_group_partial_permissions" "base_access" {
  group_id = data.dbtcloud_groups.engineering.groups[0].id
  
  permissions = [
    {
      permission_set = "member"
      all_projects   = true
    }
  ]
}

# Project team can manage project-specific permissions independently
resource "dbtcloud_scim_group_partial_permissions" "project_access" {
  group_id = data.dbtcloud_groups.engineering.groups[0].id
  
  permissions = [
    {
      permission_set = "developer"
      project_id     = data.dbtcloud_project.my_project.id
      all_projects   = false
      writable_environment_categories = ["development", "staging"]
    },
    {
      permission_set = "job_admin"
      project_id     = data.dbtcloud_project.my_project.id
      all_projects   = false
    }
  ]
}

# Example: Multiple projects managed by different teams
data "dbtcloud_project" "analytics" {
  name = "Analytics"
}

data "dbtcloud_project" "data_science" {
  name = "Data Science"
}

# Analytics team manages their own project permissions
resource "dbtcloud_scim_group_partial_permissions" "analytics" {
  group_id = data.dbtcloud_groups.engineering.groups[0].id
  
  permissions = [
    {
      permission_set = "developer"
      project_id     = data.dbtcloud_project.analytics.id
      all_projects   = false
      writable_environment_categories = ["development"]
    }
  ]
}

# Data Science team manages their own project permissions
resource "dbtcloud_scim_group_partial_permissions" "data_science" {
  group_id = data.dbtcloud_groups.engineering.groups[0].id
  
  permissions = [
    {
      permission_set = "analyst"
      project_id     = data.dbtcloud_project.data_science.id
      all_projects   = false
    }
  ]
}
