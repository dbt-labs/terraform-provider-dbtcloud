# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_license_map.my_license_map
  id = "license_map_id"
}

import {
  to = dbtcloud_license_map.my_license_map
  id = "12345"
}

# using the older import command
terraform import dbtcloud_license_map.my_license_map "license_map_id"
terraform import dbtcloud_license_map.my_license_map 12345
