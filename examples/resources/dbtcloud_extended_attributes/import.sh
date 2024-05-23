# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_extended_attributes.test_extended_attributes
  id = "project_id_id:extended_attributes_id"
}

import {
  to = dbtcloud_extended_attributes.test_extended_attributes
  id = "12345:6789"
}

# using the older import command
terraform import dbtcloud_extended_attributes.test_extended_attributes "project_id_id:extended_attributes_id"
terraform import dbtcloud_extended_attributes.test_extended_attributes 12345:6789
