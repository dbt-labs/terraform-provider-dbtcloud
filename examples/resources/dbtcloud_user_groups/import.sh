# Import using the User ID
# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_user_groups.my_user_groups
  id = "user_id"
}

import {
  to = dbtcloud_user_groups.my_user_groups
  id = "123456"
}

# using the older import command
terraform import dbtcloud_user_groups.my_user_groups "user_id"
terraform import dbtcloud_user_groups.my_user_groups 123456
