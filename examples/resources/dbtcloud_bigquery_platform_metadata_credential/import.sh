# The private key is never read back from dbt Cloud. After importing a credential that uses
# a service account key, the first apply sends the private key again. This is expected.

# using import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_bigquery_platform_metadata_credential.my_credential
  id = "credential_id"
}

import {
  to = dbtcloud_bigquery_platform_metadata_credential.my_credential
  id = "1234"
}

# using the older import command
terraform import dbtcloud_bigquery_platform_metadata_credential.my_credential "credential_id"
terraform import dbtcloud_bigquery_platform_metadata_credential.my_credential 1234
