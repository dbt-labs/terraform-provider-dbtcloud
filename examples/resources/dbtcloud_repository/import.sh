# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_repository.my_repository
  id = "project_id:repository_id"
}

import {
  to = dbtcloud_repository.my_repository
  id = "12345:6789"
}

# using the older import command
terraform import dbtcloud_repository.my_repository "project_id:repository_id"
terraform import dbtcloud_repository.my_repository 12345:6789
