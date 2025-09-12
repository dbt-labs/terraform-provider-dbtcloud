terraform {
  required_providers {
    dbtcloud = {
      source = "dbt-labs/dbtcloud"
    }
  }
}

# Get all active groups
data "dbtcloud_groups" "all_groups" {
  state = "active"
}

# Get groups with names containing "dev"
data "dbtcloud_groups" "dev_groups" {
  name_contains = "dev"
  state         = "active"
}

# Get a specific group by exact name
data "dbtcloud_groups" "specific_group" {
  name = "Developers"
}

# Create a local map of group names to IDs for easy reference
locals {
  group_map = {
    for group in data.dbtcloud_groups.all_groups.groups : group.name => group.id
  }
  
  # Filter SCIM managed groups
  scim_groups = [
    for group in data.dbtcloud_groups.all_groups.groups : group 
    if group.scim_managed == true
  ]
  
  manual_groups = [
    for group in data.dbtcloud_groups.all_groups.groups : group 
    if group.scim_managed == false
  ]
}

# Output examples
output "all_group_names" {
  value = [for group in data.dbtcloud_groups.all_groups.groups : group.name]
}

output "group_id_map" {
  value = local.group_map
}

output "scim_managed_groups" {
  value = local.scim_groups
}
