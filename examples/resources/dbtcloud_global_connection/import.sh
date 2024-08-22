# A project-scoped connection can be imported as a global connection by specifying the connection ID
# Migrating from project-scoped connections to global connections could be done by:
# 1. Adding the config for the global connection and importing it (see below)
# 2. Removing the project-scoped connection from the config AND from the state
#    - CAREFUL: If the connection is removed from the config but not the state, it will be destroyed on the next apply


# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_global_connection.my_connection
  id = "connection_id"
}

import {
  to = dbtcloud_global_connection.my_connection
  id = "1234"
}

# using the older import command
terraform import dbtcloud_global_connection.my_connection "connection_id"
terraform import dbtcloud_global_connection.my_connection 1234
