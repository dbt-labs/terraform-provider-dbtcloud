# terraform-provider-dbt-cloud
Terraform Provider for DBT Cloud

Primarily focused on managing jobs in DBT Cloud, given what
is available via the API.
Data sources for other concepts are added for convenience.
In order to use this provider, add the following to your Terraform providers
setup, with the latest version number.
```terraform
terraform {
  required_providers {
    dbt = {
      source  = "GtheSheep/dbt-cloud"
      version = "<version>"
    }
  }
}
```

## Authentication

If you want to explicitly set the authentication variables on the provider, you
can do so as below, though likely via a `variables.tf` file or config in your
CI-CD pipeline to keep these credentials safe.
```terraform
provider "dbt" {
  // required
  account_id = ...
  token      = "..."
  host_url   = "..."
}
```
You can also set them via environment variables:  
`DBT_CLOUD_ACCOUNT_ID` for the `account_id`.  
`DBT_CLOUD_TOKEN` for the `token`.  
`DBT_CLOUD_HOST_URL` (Optional) for the `host_url`.

## Examples
Check out the `examples/` folder for some usage options, these are intended to
simply showcase what this module can do rather than be best practices for any
given use case.

## Running Acceptance Tests
Currently acceptance tests, run via `make test-acceptance` must be done on your
own account, as there is no free tier of DBT Cloud that grants API access
