---
page_title: "Resource: dbtcloud_scim_group_partial_permissions"
subcategory: ""
description: |-
  Provide a partial set of permissions for an externally managed group (e.g., SCIM, manually created).
  This resource ONLY manages a subset of permissions and never creates or deletes groups.
  This is designed for federated permission management where a platform team sets global permissions
  and individual teams manage their own project-specific permissions for the same group.
  ⚠️  Important Differences:
  dbt_cloud_group: Creates group and fully manages ALL permissions (single Terraform workspace)dbt_cloud_group_partial_permissions: Creates group and manages PARTIAL permissions (multiple Terraform workspaces)dbt_cloud_scim_group_permissions: Externally-managed group, fully manages ALL permissions (replaces all permissions)dbt_cloud_scim_group_partial_permissions: Externally-managed group, manages PARTIAL permissions (adds/removes only specified permissions)
  Use Case:
  Group exists in external identity provider (e.g., Okta, Azure AD) and syncs via SCIMPlatform team manages base permissions (e.g., account-level access)Individual teams manage their own project-specific permissionsMultiple Terraform workspaces can safely manage different permissions for the same group
  ⚠️  Do not mix different resource types for the same group:
  Don't use dbt_cloud_scim_group_permissions (full permissions) with dbt_cloud_scim_group_partial_permissions (partial permissions)Don't use dbt_cloud_group or dbt_cloud_group_partial_permissions for externally managed groups
  The resource currently requires a Service Token with Account Admin access.
  Behavior:
  When creating: Adds specified permissions to the existing group (if not already present)When updating: Adds new permissions and removes old permissions from this resourceWhen deleting: Removes only the permissions managed by this resource (group and other permissions remain)
---

# Resource: dbtcloud_scim_group_partial_permissions

Provide a partial set of permissions for an externally managed group (e.g., SCIM, manually created). 
This resource ONLY manages a subset of permissions and never creates or deletes groups.

This is designed for federated permission management where a platform team sets global permissions 
and individual teams manage their own project-specific permissions for the same group.

⚠️  **Important Differences:**
- `dbt_cloud_group`: Creates group and fully manages ALL permissions (single Terraform workspace)
- `dbt_cloud_group_partial_permissions`: Creates group and manages PARTIAL permissions (multiple Terraform workspaces)
- `dbt_cloud_scim_group_permissions`: Externally-managed group, fully manages ALL permissions (replaces all permissions)
- `dbt_cloud_scim_group_partial_permissions`: Externally-managed group, manages PARTIAL permissions (adds/removes only specified permissions)

**Use Case:**
- Group exists in external identity provider (e.g., Okta, Azure AD) and syncs via SCIM
- Platform team manages base permissions (e.g., account-level access)
- Individual teams manage their own project-specific permissions
- Multiple Terraform workspaces can safely manage different permissions for the same group

⚠️  Do not mix different resource types for the same group:
- Don't use `dbt_cloud_scim_group_permissions` (full permissions) with `dbt_cloud_scim_group_partial_permissions` (partial permissions)
- Don't use `dbt_cloud_group` or `dbt_cloud_group_partial_permissions` for externally managed groups

The resource currently requires a Service Token with Account Admin access.

**Behavior:**
- When creating: Adds specified permissions to the existing group (if not already present)
- When updating: Adds new permissions and removes old permissions from this resource
- When deleting: Removes only the permissions managed by this resource (group and other permissions remain)

~> This resource is designed for **federated permission management** where multiple teams manage different permissions for the same externally-managed group (e.g., SCIM groups).

## Use Case Guidelines

Choose the right resource for your use case:

| Resource | Group Creation | Permission Management | Use When |
|----------|---------------|----------------------|----------|
| `dbtcloud_group` | ✅ Terraform creates | Full (replaces all) | Single Terraform workspace manages everything |
| `dbtcloud_group_partial_permissions` | ✅ Terraform creates | Partial (adds/removes) | Multiple workspaces manage same Terraform-created group |
| `dbtcloud_scim_group_permissions` | ❌ External (SCIM) | Full (replaces all) | External group, single workspace manages all permissions |
| `dbtcloud_scim_group_partial_permissions` | ❌ External (SCIM) | Partial (adds/removes) | External group, multiple workspaces manage different permissions |

## Federated Permission Management Pattern

This resource enables the following pattern:

```hcl
# Platform team manages base permissions (workspace 1)
data "dbtcloud_groups" "data_platform" {
  name = "data-platform-team"
}

resource "dbtcloud_scim_group_partial_permissions" "platform_base" {
  group_id = data.dbtcloud_groups.data_platform.groups[0].id
  
  permissions = [
    {
      permission_set = "member"
      all_projects   = true
    }
  ]
}

# Analytics team manages their project permissions (workspace 2)
data "dbtcloud_groups" "data_platform" {
  name = "data-platform-team"
}

data "dbtcloud_project" "analytics" {
  name = "Analytics"
}

resource "dbtcloud_scim_group_partial_permissions" "analytics_team" {
  group_id = data.dbtcloud_groups.data_platform.groups[0].id
  
  permissions = [
    {
      permission_set = "developer"
      project_id     = data.dbtcloud_project.analytics.id
      all_projects   = false
      writable_environment_categories = ["development", "staging"]
    }
  ]
}
```

Both resources manage different permissions for the same group without conflicts.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `group_id` (Number) The ID of the existing group to manage partial permissions for. This group must already exist and is typically from an external identity provider synced via SCIM.

### Optional

- `permissions` (Attributes Set) Partial set of permissions to apply to the group. These permissions will be added to any existing permissions. Other permissions on the group will not be affected. (see [below for nested schema](#nestedatt--permissions))

### Read-Only

- `id` (Number) The ID of the group (same as group_id)

<a id="nestedatt--permissions"></a>
### Nested Schema for `permissions`

Required:

- `all_projects` (Boolean) Whether access should be provided for all projects or not.
- `permission_set` (String) Set of permissions to apply. The permissions allowed are the same as the ones for the `dbtcloud_group` resource.

Optional:

- `project_id` (Number) Project ID to apply this permission to for this group.
- `writable_environment_categories` (Set of String) What types of environments to apply Write permissions to. 
Even if Write access is restricted to some environment types, the permission set will have Read access to all environments. 
The values allowed are `all`, `development`, `staging`, `production` and `other`. 
Not setting a value is the same as selecting `all`. 
Not all permission sets support environment level write settings, only `analyst`, `database_admin`, `developer`, `git_admin` and `team_admin`.

## Import

~> **Import Not Supported:** This resource does not support `terraform import` because it manages only a partial subset of permissions. 
There is no way for Terraform to know which specific permissions this resource instance should manage versus permissions 
managed by other resources or applied outside of Terraform. You must define the resource in your configuration from the start.
