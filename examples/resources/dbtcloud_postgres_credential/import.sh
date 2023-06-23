# Import using a project ID and credential ID found in the URL or via the API.
terraform import dbtcloud_postgres_credential.my_credential "project_id:credential_id"
terraform import dbtcloud_postgres_credential.my_credential 12345:6789
