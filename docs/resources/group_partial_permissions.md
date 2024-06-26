---
page_title: "dbtcloud_group_partial_permissions Resource - dbtcloud"
subcategory: ""
description: |-
  Provide a partial set of permissions for a group. This is different from dbt_cloud_group as it allows to have multiple resources updating the same dbt Cloud group and is useful for companies managing a single dbt Cloud Account configuration from different Terraform projects/workspaces.
  If a company uses only one Terraform project/workspace to manage all their dbt Cloud Account config, it is recommended to use dbt_cloud_group instead of dbt_cloud_group_partial_permissions.
  ~> This is currently an experimental resource and any feedback is welcome in the GitHub repository.
  The resource currently requires a Service Token with Account Admin access.
  The current behavior of the resource is the following:
  when using dbt_cloud_group_partial_permissions, don't use dbt_cloud_group for the same group in any other project/workspace. Otherwise, the behavior is undefined and partial permissions might be removed.when defining a new dbt_cloud_group_partial_permissions
  if the group doesn't exist with the given name, it will be createdif a group exists with the given name, permissions will be added in the dbt Cloud group if they are not present yetin a given Terraform project/workspace, avoid having different dbt_cloud_group_partial_permissions for the same group name to prevent sync issues. Add all the permissions in the same resource.all resources for the same group name need to have the same values for assign_by_default and sso_mapping_groups. Those fields are not considered "partial". (Please raise feedback in GitHub if you think that sso_mapping_groups should be "partial" as well)when a resource is updated, the dbt Cloud group will be updated accordingly, removing and adding permissionswhen the resource is deleted/destroyed, if the resulting permission sets is empty, the group will be deleted ; otherwise, the group will be updated, removing the permissions from the deleted resource
---

# dbtcloud_group_partial_permissions (Resource)


Provide a partial set of permissions for a group. This is different from `dbt_cloud_group` as it allows to have multiple resources updating the same dbt Cloud group and is useful for companies managing a single dbt Cloud Account configuration from different Terraform projects/workspaces.

If a company uses only one Terraform project/workspace to manage all their dbt Cloud Account config, it is recommended to use `dbt_cloud_group` instead of `dbt_cloud_group_partial_permissions`.

~> This is currently an experimental resource and any feedback is welcome in the GitHub repository.

The resource currently requires a Service Token with Account Admin access.

The current behavior of the resource is the following:

- when using `dbt_cloud_group_partial_permissions`, don't use `dbt_cloud_group` for the same group in any other project/workspace. Otherwise, the behavior is undefined and partial permissions might be removed.
- when defining a new `dbt_cloud_group_partial_permissions`
  - if the group doesn't exist with the given `name`, it will be created
  - if a group exists with the given `name`, permissions will be added in the dbt Cloud group if they are not present yet
- in a given Terraform project/workspace, avoid having different `dbt_cloud_group_partial_permissions` for the same group name to prevent sync issues. Add all the permissions in the same resource. 
- all resources for the same group name need to have the same values for `assign_by_default` and `sso_mapping_groups`. Those fields are not considered "partial". (Please raise feedback in GitHub if you think that `sso_mapping_groups` should be "partial" as well)
- when a resource is updated, the dbt Cloud group will be updated accordingly, removing and adding permissions
- when the resource is deleted/destroyed, if the resulting permission sets is empty, the group will be deleted ; otherwise, the group will be updated, removing the permissions from the deleted resource

## Example Usage

```terraform
// we add some permissions to the group "TF Group 1" (existing or not) to  a new project 
resource "dbtcloud_group_partial_permissions" "tf_group_1" {
	name  				= "TF Group 1"
	group_permissions = [
		{
			permission_set 	= "developer"
			project_id    	= dbtcloud_project.dbt_project.id
			all_projects  	= false
			writable_environment_categories = ["development", "staging"]
		},
		{
			permission_set 	= "git_admin"
			project_id    	= dbtcloud_project.dbt_project.id
			all_projects  	= false
		}
	]
}

// we add Admin permissions to the group "TF Group 2" (existing or not) to  a new project 
// it is possible to add more permissions to the same group name in other Terraform projects/workspaces, using another `dbtcloud_group_partial_permissions` resource
resource "dbtcloud_group_partial_permissions" "tf_group_2" {
	name  				= "TF Group 2"
	sso_mapping_groups 	= ["group2"]
	group_permissions = [
		{
			permission_set 	= "admin"
			project_id    	= dbtcloud_project.dbt_project.id
			all_projects  	= false
		}
	]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the group. This is used to identify an existing group

### Optional

- `assign_by_default` (Boolean) Whether the group will be assigned by default to users. The value needs to be the same for all partial permissions for the same group.
- `group_permissions` (Attributes Set) Partial permissions for the group. Those permissions will be added/removed when config is added/removed. (see [below for nested schema](#nestedatt--group_permissions))
- `sso_mapping_groups` (Set of String) Mapping groups from the IdP. At the moment the complete list needs to be provided in each partial permission for the same group.

### Read-Only

- `id` (Number) The ID of the group

<a id="nestedatt--group_permissions"></a>
### Nested Schema for `group_permissions`

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
