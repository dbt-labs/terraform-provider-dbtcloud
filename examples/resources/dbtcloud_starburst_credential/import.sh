# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_starburst_credential.my_starburst_credential
  id = "project_id:credential_id"
}

import {
  to = dbtcloud_starburst_credential.my_starburst_credential
  id = "12345:6789"
}

# using the older import command
terraform import dbtcloud_starburst_credential.my_starburst_credential "project_id:credential_id"
terraform import dbtcloud_starburst_credential.my_starburst_credential 12345:6789
