---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dbtcloud_users Data Source - dbtcloud"
subcategory: ""
description: |-
  Retrieve all users
---

# dbtcloud_users (Data Source)

Retrieve all users

## Example Usage

```terraform
// return all users in the dbt Cloud account
data "dbtcloud_users" "all" {
}

// we can use it to check if a user exists or not
// the dbtcloud_user datasource would fail if the user doesn't exist 
locals {
  user_details = [for user in data.dbtcloud_users.all.users : user if user.email == "example@amail.com"]
  user_exist   = length(local.user_details) == 1
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `users` (Attributes Set) Set of users with their internal ID end email (see [below for nested schema](#nestedatt--users))

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Read-Only:

- `email` (String) Email for the user
- `id` (Number) ID of the user
