# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_snowflake_credential.prod_snowflake_credential
  id = "project_id:credential_id"
}

import {
  to = dbtcloud_snowflake_credential.prod_snowflake_credential
  id = "12345:6789"
}

# using the older import command
terraform import dbtcloud_snowflake_credential.prod_snowflake_credential "project_id:credential_id"
terraform import dbtcloud_snowflake_credential.prod_snowflake_credential 12345:6789
