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