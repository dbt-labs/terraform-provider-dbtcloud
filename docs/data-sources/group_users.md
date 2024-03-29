---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dbtcloud_group_users Data Source - dbtcloud"
subcategory: ""
description: |-
  Returns a list of users assigned to a specific dbt Cloud group
---

# dbtcloud_group_users (Data Source)

Returns a list of users assigned to a specific dbt Cloud group

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

- `id` (String) The ID of this resource.
- `users` (Set of Object) List of users (map of ID and email) in the group (see [below for nested schema](#nestedatt--users))

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Read-Only:

- `email` (String)
- `id` (Number)
