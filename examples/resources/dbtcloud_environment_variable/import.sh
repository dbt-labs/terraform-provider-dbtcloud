# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_environment_variable.test_environment_variable
  id = "project_id:environment_variable_name"
}

import {
  to = dbtcloud_environment_variable.test_environment_variable
  id = "12345:DBT_ENV_VAR"
}

# using the older import command
terraform import dbtcloud_environment_variable.test_environment_variable "project_id:environment_variable_name"
terraform import dbtcloud_environment_variable.test_environment_variable 12345:DBT_ENV_VAR
