---
page_title: "dbtcloud_license_map Resource - dbtcloud"
subcategory: ""
description: |-
  
---

# dbtcloud_license_map (Resource)




## Example Usage

```terraform
# Developer license group mapping
resource "dbtcloud_license_map" "dev_license_map" {
  license_type               = "developer"
  sso_license_mapping_groups = ["DEV-SSO-GROUP"]
}

# Read-only license mapping
resource "dbtcloud_license_map" "read_only_license_map" {
  license_type               = "read_only"
  sso_license_mapping_groups = ["READ-ONLY-SSO-GROUP"]
}

# IT license mapping
resource "dbtcloud_license_map" "it_license_map" {
  license_type               = "it"
  sso_license_mapping_groups = ["IT-SSO-GROUP"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `license_type` (String) License type

### Optional

- `sso_license_mapping_groups` (Set of String) SSO license mapping group names for this group

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_license_map.my_license_map
  id = "license_map_id"
}

import {
  to = dbtcloud_license_map.my_license_map
  id = "12345"
}

# using the older import command
terraform import dbtcloud_license_map.my_license_map "license_map_id"
terraform import dbtcloud_license_map.my_license_map 12345
```
