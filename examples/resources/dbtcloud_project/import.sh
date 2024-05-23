# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_project.my_project
  id = "project_id"
}

import {
  to = dbtcloud_project.my_project
  id = "12345"
}

# using the older import command
terraform import dbtcloud_project.my_project "project_id"
terraform import dbtcloud_project.my_project 12345
