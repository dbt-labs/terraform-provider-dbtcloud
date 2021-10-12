output "test_project_name" {
  value       = data.dbt_cloud_project.test_project.name
  description = "Name of the example project"
}
