// use dbt_cloud_project instead of dbtcloud_project for the legacy resource names
// legacy names will be removed from 0.3 onwards

// projects data sources can use the project_id parameter (preferred uniqueness is ensured)
data "dbtcloud_project" "project_by_id" {
  id = 00000000000000
}

// or they can use project names
// the provider will raise an error if more than one project is found with the same name
data "dbtcloud_project" "project_by_name" {
  name = "Project name"
}

data "dbtcloud_projects" "filtered_projects" {
  name_contains = "Project"
}

data "dbtcloud_projects" "all_projects" {
}

output "project_id_details" {
  value = data.dbtcloud_project.project_by_id
}

output "project_name_details" {
  value = data.dbtcloud_project.project_by_name
}

output "filtered_projects_count" {
  value = length(data.dbtcloud_projects.filtered_projects.projects)
}

output "filtered_projects" {
  value = data.dbtcloud_projects.filtered_projects.projects
}

output "project_names" {
  value = [for project in data.dbtcloud_projects.filtered_projects.projects : project.name]
}
