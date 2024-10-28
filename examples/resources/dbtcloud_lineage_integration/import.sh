# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_lineage_integration.my_lineage_integration
  id = "projet_id:lineage_integration_id"
}

import {
  to = dbtcloud_lineage_integration.my_lineage_integration
  id = "123:4567"
}

# using the older import command
terraform import dbtcloud_lineage_integration.my_lineage_integration "projet_id:lineage_integration_id"
terraform import dbtcloud_lineage_integration.my_lineage_integration 123:4567
