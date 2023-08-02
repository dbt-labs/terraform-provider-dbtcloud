# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.2...HEAD)

## [0.2.3](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.2...v0.2.3)

## Fixes

- Update CI to avoid Node version warnings
- Fixes to the docs

## Changes

- [164](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/164) Add the ability to define `priority` and `execution_project` for BigQuery connections
- [168](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/168) Add the ability to set up email notifications (to internal users and external email addresses) based on jobs results

## [0.2.2](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.1...v0.2.2)

## Fixes

- [156](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/156) Fix the `dbtcloud_connection` for Databricks when updating the `http_path` or `catalog` + add integration test
- [157](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/157) Fix updating an environment with credentials already set + add integration test

## Changes

- Add guide to get started with the provider
- Add missing import and fix more docs
- Update docs template to allow using Subcategories later

## [0.2.1](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.0...v0.2.1)

## Changes

- Resources deleted from dbt Cloud won't crash the provider and we now consider the resource as deleted, removing it from the state. This is the expected behaviour of a provider.
- Add examples in the docs to resources that didn't have any so far

## [0.2.0](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.1.12...v0.2.0)

## Important changes

- The resources and data sources are now available as `dbtcloud_xxx` (following the terraform convention) in addition to `dbt_cloud_xxx` (legacy). The legacy version will be removed from v0.3.0 onwards. Instructions on how to use the new resources are available on [the main page of the Provider](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs).

## 0.1.12

## Changes

- The provider is now published under the dbt-labs org: https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest
