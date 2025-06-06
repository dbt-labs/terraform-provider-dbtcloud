---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dbtcloud_group_users Data Source - dbtcloud"
subcategory: ""
description: |-
  Databricks credential data source
---

# dbtcloud_group_users (Data Source)

Databricks credential data source

## Example Usage

```terraform
data "dbtcloud_group_users" "my_group_users" {
  group_id = 1234
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `group_id` (Number) ID of the group

### Read-Only

- `id` (String) The ID of this resource. Contains the project ID and the credential ID.
- `users` (Attributes Set) List of users (map of ID and email) in the group (see [below for nested schema](#nestedatt--users))

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Required:

- `email` (String) Email of the user
- `id` (Number) ID of the user
