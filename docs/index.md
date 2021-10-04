---
page_title: "Provider: dbt-cloud"
description: Manage DBT Cloud with Terraform.
---

# DBT Cloud Provider

This is a terraform provider plugin for managing [DBT Cloud](https://cloud.getdbt.com/) accounts.
Given the current capabilities of the API we focus on the management of job definitions.
Also, the API doesn't support deletion, we can only set the state to deleted.
## Example Provider Configuration

```terraform
provider "dbt" {
  // required
  account_id = ...
  token      = "..."
}
```

### Required

- **account_id** (Integer)
- **token** (String)
