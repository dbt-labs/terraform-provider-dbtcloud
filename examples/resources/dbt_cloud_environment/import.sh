# Import using a project ID and environment ID found in the URL or via the API.
terraform import dbt_cloud_environment.test_environment "project_id:environment_id"
terraform import dbt_cloud_environment.test_environment 12345:6789
