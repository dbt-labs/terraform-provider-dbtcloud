# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_spark_credential.my_spark_credential
  id = "project_id:credential_id"
}

import {
  to = dbtcloud_spark_credential.my_spark_credential
  id = "12345:6789"
}

# using the older import command
terraform import dbtcloud_spark_credential.my_spark_credential "project_id:credential_id"
terraform import dbtcloud_spark_credential.my_spark_credential 12345:6789

