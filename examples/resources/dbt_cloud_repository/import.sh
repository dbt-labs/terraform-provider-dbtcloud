# Import using a project ID and repository ID found in the URL or via the API.
terraform import dbt_cloud_repository.test_repository "project_id:repository_id"
terraform import dbt_cloud_repository.test_repository 12345:6789
