# Import using a project ID and Connection ID found in the URL or via the API.
terraform import dbtcloud_project_repository.my_project "project_id:repository_id"
terraform import dbtcloud_project_repository.my_project 12345:5678
