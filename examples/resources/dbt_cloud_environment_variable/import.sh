# Import using a project ID and environment variable name found in the URL and UI or via the API.
terraform import dbt_cloud_environment_variable.test_environment_variable "project_id:environment_variable_name"
terraform import dbt_cloud_environment_variable.test_environment_variable 12345:DBT_ENV_VAR
