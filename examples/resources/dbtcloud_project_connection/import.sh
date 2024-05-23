# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_project_connection.my_project
  id = "project_id:connection_id"
}

import {
  to = dbtcloud_project_connection.my_project
  id = "12345:5678"
}

# using the older import command
terraform import dbtcloud_project_connection.my_project "project_id:connection_id"
terraform import dbtcloud_project_connection.my_project 12345:5678
