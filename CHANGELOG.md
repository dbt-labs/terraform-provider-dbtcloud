# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.3.12...HEAD)

# [0.3.12](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.3.11...v0.3.12)

### Changes

- Add support for `import` for `dbtcloud_global_connection`
- Add support for Databricks in `dbtcloud_global_connection` 

# [0.3.11](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.3.10...v0.3.11)

### Changes

- [#267](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/267) Support for global connections
  - `dbtcloud_environment` now accepts a `connection_id` to link the environment to the connection. This is the new recommended way to link connections to environments instead of linking the connection to the project with `dbtcloud_project_connection`
    - The `dbtcloud_project_connection` still works today and when used doesn't require setting up a `connection_id` in the `dbtcloud_environment` resource (i.e. , any current config/module should continue working), but the resource is flagged as deprecated and will be removed in a future version of the provider
  - For now, people can continue using the project-scoped connection resources `dbtcloud_connection`, `dbtcloud_bigquery_connection` and `dbtcloud_fabric_connection` for creating and updating global connections. The parameter `project_id` in those connections still need to be a valid project id but doesn't mean that this connection is restricted to this project ID. The project-scoped connections created from Terraform are automatically converted to global connections
  - A new resource `dbtcloud_global_connection` has been created and currently supports Snowflake and BigQuery connections. In the next weeks, support for all the Data Warehouses will be added to this resource
    - When a data warehouse is supported in `dbtcloud_global_connection`, we recommend using this new resource instead of the legacy project-scoped connection resources. Those resources will be deprecated in a future version of the provider.
- [#278](https://github.com/dbt-labs/terraform-provider-dbtcloud/pull/278) Deprecate `state` attribute in the resources and datasources that use it. It will be removed in the next major version of the provider. This attribute is used for soft-delete and isn't intended to be configured in the scope of the provider.

### Fix

- [#281](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/281) Fix the datasource `dbcloud_environments` where the environment IDs were not being saved

# [0.3.10](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.3.9...v0.3.10)

### Changes

- [#277](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/277) Add `dbtcloud_users` datasource to get all users
- [#274](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/274) Add `dbtcloud_jobs` datasource to return all jobs for a given dbt Cloud project or environment
- [#273](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/273) Add environment level restrictions to the `dbtcloud_service_token` resource

### Docs

- Fix typo in service token examples

## [0.3.9](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.3.8...v0.3.9)

### Fixes

- [#271](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/271) Force creation of a new connection when the project is changed or deleted

### Docs

- Fix typo in environment code example

## [0.3.8](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.3.7...v0.3.8)

### Changes

- Added new `on_warning` field to `dbtcloud_notification` and `dbtcloud_partial_notification`. 

## [0.3.7](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.3.6...v0.3.7)

### Changes

- [#266](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/266) Add env level permissions for `dbtcloud_group` and `dbtcloud_group_partial_permissions`. As of June 5 this feature is not yet active for all customers.

### Docs

- Fix description of fields for some datasources

### Internals

- Move the `dbcloud_group` resource and datasource from the SDKv2 to the Framework
- Create new helpers for comparing Go structs
- Update all SDKv2 tests to run on the muxed provider to work when some resources have moved to the Plugin Framework

## [0.3.6](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.3.5...v0.3.6)

### Changes

- [#232](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/232) add deprecation notice for `dbtcloud_project_artefacts` as the resource is not required now that dbt Explorer is GA.
- [#208](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/208) add new `dbtcloud_partial_license_map` for defining SSO group mapping to license types from different Terraform projects/resources

## [0.3.5](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.3.4...v0.3.5)

### Changes

- add a `dbtcloud_partial_notification` resource to allow different resources to add/remove job notifications for the same Slack channel/email/user

### Fixes

- [#257](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/257) - Force new resource when the `project_id` changes for a `dbtcloud_job`.
- Creating connection for adapters (e.g. Databricks and Fabric) was failing when using Service Tokens following changes in the dbt Cloud APIs

### Behind the scenes

- change the User Agent to report what provider version is being used

### Documentation

- add import block example for the resources in addition to the import command

## [0.3.4](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.3.3...v0.3.4)

### Changes

- [#255](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/255) - Add new datasource `dbtcloud_environments` to return all environments across an account, or all environments for a give project ID

### Behind the scenes

- Move the `dbtcloud_environment` datasource to the Terraform Plugin Framework

## [0.3.3](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.3.2...v0.3.3)

### Changes

- [#250](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/250) - [Experimental] Create a new resource called `dbtcloud_group_partial_permissions` to manage permissions of a single group from different resources which can be set across different Terraform projects/workspaces. The dbt Cloud API doesn't provide endpoints for adding/removing single permissions, so the logic in the provider is more complex than other resources. If the resource works as expected for the provider users we could create similar ones for "partial" notifications and "partial" license mappings.

## [0.3.2](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.3.0...v0.3.2)

### Changes

- Add `on_merge` trigger for jobs. The trigger is optional for now but will be required in the future. 

### Documentation

- Remove mention of `dbt_cloud_xxx` resources in the docs

## [0.3.0](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.25...v0.3.0)

### Changes

- Implements muxing to allow both SDKv2 and Plugin Framework resources to work at the same time. This change a bit the internals but shouldn't have any regression.
- Move some resources / datasources to the plugin Framework
- Remove legacy `dbt_cloud_xxx` resources

## [0.2.25](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.24...v0.2.25)

### Changes

- Enable OAuth configuration for Databricks connections + update docs accordingly

## [0.2.24](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.23...v0.2.24)

### Fixes

- [#247](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/247) Segfault when the env var for the token is empty
- [Internal] Issue with `job_ids` required to be set going forward, even if it is empty

## [0.2.23](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.22...v0.2.23)

### Changes

- [#244](https://github.com/dbt-labs/terraform-provider-dbtcloud/pull/244) Better error handling when GitLab repositories are created with a User Token

### Fixes

- [#245](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/245) Issues on `dbtcloud_job` when modifying an existing job schedule

## [0.2.22](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.21...v0.2.22)

### Changes

- [#240](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/240) Add notice of deprecation for `triggers.custom_branch_only` for jobs and update logic to make it work even though people have it to true or false in their config. We might raise an error if the field is still there in the future.
- Update diff calculation for Extended Attributes, allowing strings which are not set with `jsonencode()`
- [#241](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/241) Force recreation of env vars when values change to work with the recent changes in the dbt Cloud API

### Documentation

- Add list of permission names and permission codes in the docs of the `service_token` and `group`
- Add info in `dbtcloud_repository` about the need to also create a `dbtcloud_project_repository`

## [0.2.21](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.20...v0.2.21)

### Changes

- Flag `fetch_deploy_key` as deprecated for `dbtcloud_repository`. The key is always fetched for the genetic git clone approach

### Documentations

- Add info about `versionless` dbt environment (Private Beta)
- [#235](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/235) Fix docs on the examples for Fabric credentials

## [0.2.20](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.19...v0.2.20)

### Changes

- Add support for job chaining and `job_completion_trigger_condition` (feature is in closed Beta in dbt Cloud as of 5 FEB 2024)

### Documentations

- Improve docs for jobs

## [0.2.19](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.18...v0.2.19)

### Changes

- Update permissions allowed for groups and token to include `job_runner`

### Documentations

- Add guide on `dbtcloud-terraforming` to import existing resources

## [0.2.18](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.17...v0.2.18)

### Changes

- #229 - fix logic for secret environment variables

### Documentations

- #228 - update docs to replace the non existing `dbtcloud_user` resource by the existing `data.dbtcloud_user` data source

### Behind the scenes

- update third party module version following security report

## [0.2.17](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.16...v0.2.17)

### Changes

- #224 - add the resources `dbtcloud_fabric_connection` and `dbtcloud_fabric_credential` to allow using dbt Cloud along with Microsoft Fabric
- #222 - allow users to set Slack notifications from Terraform

### Behind the scenes

- Refactor some of the shared code for Adapters and connections

## [0.2.16](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.15...v0.2.16)

### Changes

- #99 - add the resource `environment_variable_job_override` to allow environment variable override in jobs
- Update the go version and packages versions

### Fixes

- #221 - removing the value for an env var scope was not removing it in dbt Cloud
- add better messages and error handling for jobs

## [0.2.15](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.14...v0.2.15)

### Changes

- Update list of permissions for groups and service tokens

## [0.2.14](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.13...v0.2.14)

### Changes

- Fix issues with the repositories connected via GitLab native integration
- Add ability to configure repositories using the native ADO integration
- Add data sources for retrieving ADO projects and repositories ID and information

### Documentation

- Show in the main page that provider parameters can be set with env vars
- Update examples and field descriptions for the repositories

## [0.2.13](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.11...v0.2.13)

### Changes

- Update connections to force new one when the project changes
- Add support for the Datasource dbtcloud_group_users to get the list of users assigned to a given project

### Documentation

- Use d2 for showing the different resources
- Update examples in docs

## [0.2.11](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.10...v0.2.11)

### Changes

- Update docs and examples for jobs and add the ability to set/unset running CI jobs on Draft PRs

## [0.2.10](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.9...v0.2.10)

### Fix

- [#197](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/197) - Community contribution to handle cases where more than 100 groups are created in dbt Cloud
- [#199](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/199) - Update logic to allow finding users by their email addresses in a cases insensitive way
- [#198](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/198) - Update some internal logic to call endpoints by their unique IDs instead of looping through answers to avoid issues like #199 and paginate through results for endpoints where we can't query the ID directly

### Changes

- [#189](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/189) - Allow users to retrieve project data sources by providing project names instead of project IDs. This will return an error if more than 1 project has the given name and takes care of the pagination required for handling more than 100 projects

## [0.2.9](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.8...v0.2.9)

### Changes

- Add support for extended attributes for environments [(docs)](https://docs.getdbt.com/docs/dbt-cloud-environments#extended-attributes-beta), allowing people to add connection attributes available in dbt-core but not in the dbt Cloud interface
- [#191](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/191) - Allow setting a description for jobs

## [0.2.8](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.7...v0.2.8)

### Fix

- [#190](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/190) - Allow setting deferral for jobs at the environment level rather than at the job level. This is due to changes in CI in dbt Cloud. Add docs about those changes on the dbtcloud_job resource page

## [0.2.7](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.6...v0.2.7)

### Fix

- [#184](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/184) - Fix issue when updating SSO groups for a given RBAC group

## [0.2.6](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.5...v0.2.6)

### Changes

- [#178](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/178) and [#179](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/179): Add support for [dbtcloud_license_map](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/license_map), allowing the assignment of SSO groups to different dbt Cloud license types

## [0.2.5](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.4...v0.2.5)

### Fixes

- [#172](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/172): Fix issue when changing the schedule of jobs from a list of hours to an interval in a [dbtcloud_job](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/job)
- [#175](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/175): Fix issue when modifying the `environment_id` of an existing [dbtcloud_job](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/job)
- [#154](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/154): Allow the creation of [Databricks connections](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/connection) using Service Tokens when it was only possible with User Tokens before

### Changes

- Use the `v2/users/<id>` endpoint to get the groups of a user

## [0.2.4](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.3...v0.2.4)

### Fixes

- More update to docs

### Changes

- [#171](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/171) Add the ability to define which [environment](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/environment) is the production one (to be used with cross project references in dbt Cloud)
- Add [guide](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/guides/2_leveraging_http_provider) on how to use the Hashicorp HTTP provider
- [#174](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/174) Add the ability to assign User groups to dbt Cloud users.

## [0.2.3](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.2...v0.2.3)

### Fixes

- Update CI to avoid Node version warnings
- Fixes to the docs

### Changes

- [164](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/164) Add the ability to define `priority` and `execution_project` for [BigQuery connections](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/bigquery_connection)
- [168](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/168) Add the ability to set up [email notifications](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/resources/notification) (to internal users and external email addresses) based on jobs results

## [0.2.2](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.1...v0.2.2)

### Fixes

- [#156](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/156) Fix the `dbtcloud_connection` for Databricks when updating the `http_path` or `catalog` + add integration test
- [#157](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/157) Fix updating an environment with credentials already set + add integration test

### Changes

- Add [guide](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs/guides/1_getting_started) to get started with the provider
- Add missing import and fix more docs
- Update docs template to allow using Subcategories later

## [0.2.1](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.2.0...v0.2.1)

### Changes

- Resources deleted from dbt Cloud won't crash the provider and we now consider the resource as deleted, removing it from the state. This is the expected behavior of a provider.
- Add examples in the docs to resources that didn't have any so far

## [0.2.0](https://github.com/dbt-labs/terraform-provider-dbtcloud/compare/v0.1.12...v0.2.0)

### Important changes

- The resources and data sources are now available as `dbtcloud_xxx` (following the terraform convention) in addition to `dbt_cloud_xxx` (legacy). The legacy version will be removed from v0.3.0 onwards. Instructions on how to use the new resources are available on [the main page of the Provider](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs).

## 0.1.12

### Changes

- The provider is now published under the dbt-labs org: https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest
