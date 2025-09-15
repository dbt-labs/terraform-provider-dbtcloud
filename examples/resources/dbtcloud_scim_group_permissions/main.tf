terraform {
  required_providers {
    dbtcloud = {
      source = "dbt-labs/dbtcloud"
    }
  }
}

# Get existing groups (including SCIM-managed ones)
data "dbtcloud_groups" "existing" {
  state = "active"
}

# Example project (you would typically reference an existing project)
resource "dbtcloud_project" "example" {
  name = "Example Project for SCIM Groups"
}

# Manage permissions for a SCIM-managed group using group ID lookup
resource "dbtcloud_scim_group_permissions" "developers" {
  group_id = [for group in data.dbtcloud_groups.existing.groups : group.id if group.name == "SCIM-Developers"][0]
  
  permissions = [
    {
      permission_set                  = "developer"
      project_id                      = dbtcloud_project.example.id
      all_projects                    = false
      writable_environment_categories = ["development"]
    },
    {
      permission_set                  = "analyst"
      project_id                      = dbtcloud_project.example.id
      all_projects                    = false
      writable_environment_categories = ["development", "staging"]
    }
  ]
}

# Manage permissions for multiple SCIM groups using a local map
locals {
  scim_group_map = {
    for group in data.dbtcloud_groups.existing.groups : group.name => group.id
    if group.scim_managed == true
  }
}

resource "dbtcloud_scim_group_permissions" "data_engineers" {
  group_id = local.scim_group_map["Data Engineers"]
  
  permissions = [
    {
      permission_set                  = "developer"
      all_projects                    = true
      writable_environment_categories = ["development", "staging"]
    }
  ]
}

resource "dbtcloud_scim_group_permissions" "analysts" {
  group_id = local.scim_group_map["Business Analysts"]
  
  permissions = [
    {
      permission_set                  = "analyst"
      all_projects                    = true
      writable_environment_categories = ["development"]
    }
  ]
}

# Example with direct group ID (if you know the ID)
resource "dbtcloud_scim_group_permissions" "admin_group" {
  group_id = 12345  # Replace with actual group ID
  
  permissions = [
    {
      permission_set = "account_admin"
      all_projects   = true
    }
  ]
}
