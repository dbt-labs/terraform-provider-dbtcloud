---
page_title: "dbtcloud_databricks_credential Resource - dbtcloud"
subcategory: ""
description: |-
  Databricks credential resource
---

# dbtcloud_databricks_credential (Resource)


Databricks credential resource

## Example Usage

```terraform
resource "dbtcloud_databricks_credential" "my_databricks_cred" {
  project_id   = dbtcloud_project.dbt_project.id
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "databricks"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project_id` (Number) Project ID to create the Databricks credential in
- `token` (String, Sensitive) Token for Databricks user

### Optional

- `adapter_type` (String) The type of the adapter (databricks or spark). Optional only when semantic_layer_credential is set to true; otherwise, this field is required.
- `catalog` (String) The catalog where to create models (only for the databricks adapter)
- `schema` (String) The schema where to create models. Optional only when semantic_layer_credential is set to true; otherwise, this field is required.
- `semantic_layer_credential` (Boolean) This field indicates that the credential is used as part of the Semantic Layer configuration. It is used to create a Databricks credential for the Semantic Layer.
- `target_name` (String, Deprecated) Target name

### Read-Only

- `credential_id` (Number) The system Databricks credential ID
- `id` (String) The ID of this resource. Contains the project ID and the credential ID.

## Import

Import is supported using the following syntax:

```shell
# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_databricks_credential.my_databricks_credential
  id = "project_id:credential_id"
}

import {
  to = dbtcloud_databricks_credential.my_databricks_credential
  id = "12345:6789"
}

# using the older import command
terraform import dbtcloud_databricks_credential.my_databricks_credential "project_id:credential_id"
terraform import dbtcloud_databricks_credential.my_databricks_credential 12345:6789
```
