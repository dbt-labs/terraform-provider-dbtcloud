---
page_title: "dbtcloud_extended_attributes Resource - dbtcloud"
subcategory: ""
description: |-
  Extended attributes resource
---

# dbtcloud_extended_attributes (Resource)


Extended attributes resource

## Example Usage

```terraform
# extended_attributes can be set as a raw JSON string or encoded with Terraform's `jsonencode()` function
# we recommend using `jsonencode()` to avoid Terraform reporting changes due to whitespaces or keys ordering
resource "dbtcloud_extended_attributes" "my_attributes" {
  extended_attributes = jsonencode(
    {
      type      = "databricks"
      catalog   = "dbt_catalog"
      http_path = "/sql/your/http/path"
      my_nested_field = {
        subfield = "my_value"
      }
    }
  )
  project_id = var.dbt_project.id
}

resource "dbtcloud_environment" "issue_depl" {
  dbt_version            = "latest"
  name                   = "My environment"
  project_id             = var.dbt_project.id
  type                   = "deployment"
  use_custom_branch      = false
  credential_id          = var.dbt_credential_id
  deployment_type        = "production"
  extended_attributes_id = dbtcloud_extended_attributes.my_attributes.extended_attributes_id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `extended_attributes` (String) A JSON string listing the extended attributes mapping. The keys are the connections attributes available in the `profiles.yml` for a given adapter. Any fields entered will override connection details or credentials set on the environment or project. To avoid incorrect Terraform diffs, it is recommended to create this string using `jsonencode` in your Terraform code. (see example)
- `project_id` (Number) Project ID to create the extended attributes in

### Optional

- `state` (Number) The state of the extended attributes (1 = active, 2 = inactive)

### Read-Only

- `extended_attributes_id` (Number) Extended attributes ID
- `id` (String) The ID of this resource. Contains the project ID and the extended attributes ID.

## Import

Import is supported using the following syntax:

```shell
# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_extended_attributes.test_extended_attributes
  id = "project_id_id:extended_attributes_id"
}

import {
  to = dbtcloud_extended_attributes.test_extended_attributes
  id = "12345:6789"
}

# using the older import command
terraform import dbtcloud_extended_attributes.test_extended_attributes "project_id_id:extended_attributes_id"
terraform import dbtcloud_extended_attributes.test_extended_attributes 12345:6789
```
