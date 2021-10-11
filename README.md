# terraform-provider-dbt-cloud
Terraform Provider for DBT Cloud

Primarily focused on managing jobs in DBT Cloud, given what
is available via the API.
Data sources for other concepts are added for convenience.

```terraform
provider "dbt" {
  // required
  account_id = ...
  token      = "..."
}
```
