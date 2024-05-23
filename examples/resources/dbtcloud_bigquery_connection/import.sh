# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_bigquery_connection.my_connection
  id = "project_id:connection_id"
}

import {
  to = dbtcloud_bigquery_connection.my_connection
  id = "12345:6789"
}

# using the older import command
terraform import dbtcloud_bigquery_connection.my_connection "project_id:connection_id"
terraform import dbtcloud_bigquery_connection.my_connection 12345:6789