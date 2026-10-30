# using import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_analytics_credential.my_credential
  id = "credential_id"
}

import {
  to = dbtcloud_analytics_credential.my_credential
  id = "account_id:credential_id"
}

import {
  to = dbtcloud_analytics_credential.my_credential
  id = "1234"
}

# using the older import command
terraform import dbtcloud_analytics_credential.my_credential "credential_id"
terraform import dbtcloud_analytics_credential.my_credential "account_id:credential_id"
terraform import dbtcloud_analytics_credential.my_credential 1234

# secrets are write-only, so after an import the first plan shows an update
# that sets the `_wo_version` attributes
