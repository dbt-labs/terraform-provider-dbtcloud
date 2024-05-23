# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_group.my_group
  id = "group_id"
}

import {
  to = dbtcloud_group.my_group
  id = "12345"
}

# using the older import command
terraform import dbtcloud_group.my_group "group_id"
terraform import dbtcloud_group.my_group 12345
