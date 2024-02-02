# terraform-provider-dbtcloud

## Terraform Provider for dbt Cloud

This repo was originally created by a dbt community member, Gary James [[GtheSheep](https://github.com/GtheSheep)]

> If you are still using the GtheSheep/dbt-cloud source, see [Upgrading from the community Provider](UPGRADING_PROVIDER.md) to upgrade to the latest version.

## Scope

Provide the ability to manage dbt Cloud projects and account settings via Terraform resources.
Data sources are also available for most resources.

In order to use this provider, add the following to your Terraform providers
setup, with the latest version number.

```terraform
terraform {
  required_providers {
    dbtcloud = {
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
provider "dbtcloud" {
  // required
  account_id = ...
  token      = "..."
  // optional - defaults to the US Multi Tenant URL "https://cloud.getdbt.com/api"
  host_url   = "..."
}
```

You can also set them via environment variables:  
`DBT_CLOUD_ACCOUNT_ID` for the `account_id`.  
`DBT_CLOUD_TOKEN` for the `token`.  
`DBT_CLOUD_HOST_URL` (Optional) for the `host_url`.

## Getting started and Examples

The provider documentation is directly available [on the Terraform Registry](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs).

- Under [Guides](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/guides/1_getting_started), you will find a simple example of how to use the provider
- Each resource ([example for jobs](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/job)) has some usage examples and contains the list of parameters available

## Importing existing dbt Cloud configuration

The CLI [dbtcloud-terraforming](https://github.com/dbt-labs/dbtcloud-terraforming) can be used to generate the Terraform configuration and import statements based on your existing dbt Cloud configuration.

## Running Acceptance Tests

Currently, acceptance tests, run via `make test-acceptance` must be done on your
own account

## Acknowledgement

Thanks to Gary James [[GtheSheep](https://github.com/GtheSheep)], for all the effort put in creating this provider originally
and for being a great dbt community member!
