# Import using a project ID and repository ID found in the URL or via the API.
terraform import dbtcloud_repository.test_repository "project_id:repository_id"
terraform import dbtcloud_repository.test_repository 12345:6789
