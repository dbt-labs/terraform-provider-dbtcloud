# using import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_azure_ad_application.this
  id = "azure_ad_application_id"
}

import {
  to = dbtcloud_azure_ad_application.this
  id = "12345"
}

# using the older import command
terraform import dbtcloud_azure_ad_application.this azure_ad_application_id
terraform import dbtcloud_azure_ad_application.this 12345

# NOTE: client_id, client_secret, and tenant_id will be empty after import —
# the API never returns these values. You must set them in your config to
# avoid drift on the next apply.
#
# Import is also the recovery path if destroy left an orphaned record in dbt
# Cloud (the DELETE endpoint soft-deletes the row rather than removing it).
# Find the existing record ID and import it instead of creating a new one.
