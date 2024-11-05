# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_oauth_configuration.my_external_oauth
  id = "external_oauth_id"
}

import {
  to = dbtcloud_oauth_configuration.my_external_oauth
  id = "12345"
}

# using the older import command
terraform import dbtcloud_oauth_configuration.my_external_oauth "external_oauth_id"
terraform import dbtcloud_oauth_configuration.my_external_oauth 12345
