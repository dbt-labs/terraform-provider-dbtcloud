# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_environment_variable_job_override.test_environment_variable_job_override
  id = "project_id:job_id:environment_variable_override_id"
}

import {
  to = dbtcloud_environment_variable_job_override.test_environment_variable_job_override
  id = "12345:678:123456"
}

# using the older import command
terraform import dbtcloud_environment_variable_job_override.test_environment_variable_job_override "project_id:job_id:environment_variable_override_id"
terraform import dbtcloud_environment_variable_job_override.test_environment_variable_job_override 12345:678:123456
