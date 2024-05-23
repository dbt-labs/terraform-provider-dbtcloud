# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_environment.prod_environment
  id = "project_id:environment_id"
}

import {
  to = dbtcloud_environment.prod_environment
  id = "12345:6789"
}

# using the older import command
terraform import dbtcloud_environment.prod_environment "project_id:environment_id"
terraform import dbtcloud_environment.prod_environment 12345:6789
