# Import using a project ID and environment ID found in the URL or via the API.
terraform import dbtcloud_environment.prod_environment "project_id:environment_id"
terraform import dbtcloud_environment.prod_environment 12345:6789
