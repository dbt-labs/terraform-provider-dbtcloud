# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_profile.my_profile
  id = "project_id:profile_id"
}

import {
  to = dbtcloud_profile.my_profile
  id = "12345:6789"
}

# using the older import command
terraform import dbtcloud_profile.my_profile "project_id:profile_id"
terraform import dbtcloud_profile.my_profile 12345:6789
