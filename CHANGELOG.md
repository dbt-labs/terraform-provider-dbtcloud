# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.5...HEAD)

## [0.2.5](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.4...v0.2.5)

## Fixes

- [#172](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/172): Fix issue when changing the schedule of jobs from a list of hours to an interval in a [dbtcloud_job](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/job)
- [#175](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/175): Fix issue when modifying the `environment_id` of an existing [dbtcloud_job](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/job)
- [#154](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/154): Allow the creation of [Databricks connections](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/connection) using Service Tokens when it was only possible with User Tokens before

## Changes

- Use the `v2/users/<id>` endpoint to get the groups of a user

## [0.2.4](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.3...v0.2.4)

## Fixes

- More update to docs

## Changes

- [#171](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/171) Add the ability to define which [environment](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/environment) is the production one (to be used with cross project references in dbt Cloud)
- Add [guide](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/guides/2_leveraging_http_provider) on how to use the Hashicorp HTTP provider
- [#174](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/174) Add the ability to assign User groups to dbt Cloud users.

## [0.2.3](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.2...v0.2.3)

## Fixes

- Update CI to avoid Node version warnings
- Fixes to the docs

## Changes

- [164](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/164) Add the ability to define `priority` and `execution_project` for [BigQuery connections](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/bigquery_connection)
- [168](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/168) Add the ability to set up [email notifications](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/notification) (to internal users and external email addresses) based on jobs results

## [0.2.2](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.1...v0.2.2)

## Fixes

- [#156](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/156) Fix the `dbtcloud_connection` for Databricks when updating the `http_path` or `catalog` + add integration test
- [#157](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/157) Fix updating an environment with credentials already set + add integration test

## Changes

- Add [guide](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/guides/1_getting_started) to get started with the provider
- Add missing import and fix more docs
- Update docs template to allow using Subcategories later

## [0.2.1](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.0...v0.2.1)

## Changes

- Resources deleted from dbt Cloud won't crash the provider and we now consider the resource as deleted, removing it from the state. This is the expected behavior of a provider.
- Add examples in the docs to resources that didn't have any so far

## [0.2.0](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.1.12...v0.2.0)

## Important changes

- The resources and data sources are now available as `dbtcloud_xxx` (following the terraform convention) in addition to `dbt_cloud_xxx` (legacy). The legacy version will be removed from v0.3.0 onwards. Instructions on how to use the new resources are available on [the main page of the Provider](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs).

## 0.1.12

## Changes

- The provider is now published under the dbt-labs org: https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest
