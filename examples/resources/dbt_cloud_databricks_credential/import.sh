# Import using a project ID and credential ID found in the URL or via the API.
terraform import dbt_cloud_databricks_credential.test_databricks_credential "project_id:credential_id"
terraform import dbt_cloud_databricks_credential.test_databricks_credential 12345:6789
