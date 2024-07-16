# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_service_token.my_service_token
  id = "service_token_id"
}

import {
  to = dbtcloud_service_token.my_service_token
  id = "12345"
}

# using the older import command
terraform import dbtcloud_service_token.my_service_token "service_token_id"
terraform import dbtcloud_service_token.my_service_token 12345
