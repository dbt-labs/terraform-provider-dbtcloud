# terraform-provider-dbtcloud

## Terraform Provider for dbt Cloud

This repo was originally created by a dbt community member, Gary James [[GtheSheep](https://github.com/GtheSheep)]

| If you are still using the GtheSheep/dbt-cloud source, see [Upgrading from the community Provider](UPGRADING_PROVIDER.md) to upgrade to the latest version.

## Scope

Provide the ability to manage dbt Cloud projects and account settings via Terraform resources.
Data sources are also available for most resources.

In order to use this provider, add the following to your Terraform providers
setup, with the latest version number.

```terraform
terraform {
  required_providers {
    dbt = {
      source  = "dbt-labs/dbtcloud"
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

Currently, acceptance tests, run via `make test-acceptance` must be done on your
own account

## Acknowledgement

Thanks to Gary James [[GtheSheep](https://github.com/GtheSheep)], for all the effort put in creating this provier originally
and for being a great dbt community member!
