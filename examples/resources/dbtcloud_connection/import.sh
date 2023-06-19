# Import using a project ID and connection ID found in the URL or via the API.
terraform import dbtcloud_connection.test_connection "project_id:connection_id"
terraform import dbtcloud_connection.test_connection 12345:6789
