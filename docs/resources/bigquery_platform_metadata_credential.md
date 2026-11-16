---
page_title: "dbtcloud_bigquery_platform_metadata_credential Resource - dbtcloud"
subcategory: ""
description: |-
  Manages BigQuery platform metadata credentials for external metadata ingestion in dbt Cloud.
  This resource configures credentials that allow dbt Cloud to connect directly to your BigQuery project
  to ingest metadata outside of normal dbt project runs. This enables features like:
  Catalog Ingestion: Ingest metadata about tables/views not defined in dbtCost Optimization: Query warehouse cost and performance dataCost Insights: Enhanced cost visibility and analysis
  Two authentication types are supported:
  service-account-json (default): the fields of a service account JSON key. Works with
  bigquery_v0 and bigquery_v1 connections.external-oauth-wif: Workload Identity Federation. Requires a bigquery_v1 connection
  (use_latest_adapter = true on dbtcloud_global_connection).
  The private key can be rotated and auth_type can be changed in place; neither recreates the credential.
  ~> Note: At least one of catalog_ingestion_enabled, cost_optimization_enabled, or
  cost_insights_enabled must be enabled for the credential to be usable.
  ~> Note: The connection_id cannot be changed after creation. To use a different connection,
  you must destroy and recreate the resource.
  ~> Note: The private key is never read back from dbt Cloud. After importing a credential that uses
  service-account-json, the first apply sends the private key again. This is expected.
---

# dbtcloud_bigquery_platform_metadata_credential (Resource)


Manages BigQuery platform metadata credentials for external metadata ingestion in dbt Cloud.

This resource configures credentials that allow dbt Cloud to connect directly to your BigQuery project
to ingest metadata outside of normal dbt project runs. This enables features like:

- **Catalog Ingestion**: Ingest metadata about tables/views not defined in dbt
- **Cost Optimization**: Query warehouse cost and performance data
- **Cost Insights**: Enhanced cost visibility and analysis

Two authentication types are supported:

- `service-account-json` (default): the fields of a service account JSON key. Works with
`bigquery_v0` and `bigquery_v1` connections.
- `external-oauth-wif`: Workload Identity Federation. Requires a `bigquery_v1` connection
(`use_latest_adapter = true` on `dbtcloud_global_connection`).

The private key can be rotated and `auth_type` can be changed in place; neither recreates the credential.

~> **Note:** At least one of `catalog_ingestion_enabled`, `cost_optimization_enabled`, or
`cost_insights_enabled` must be enabled for the credential to be usable.

~> **Note:** The `connection_id` cannot be changed after creation. To use a different connection,
you must destroy and recreate the resource.

~> **Note:** The private key is never read back from dbt Cloud. After importing a credential that uses
`service-account-json`, the first apply sends the private key again. This is expected.

## Example Usage

```terraform
// Service account key, with the write-only attribute (not stored in state, requires Terraform >= 1.11).
// The connection can use bigquery_v0 or bigquery_v1.
variable "bigquery_metadata_private_key" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_bigquery_platform_metadata_credential" "example" {
  connection_id = dbtcloud_global_connection.bigquery.id

  catalog_ingestion_enabled = true
  cost_optimization_enabled = false
  cost_insights_enabled     = true

  // to rotate the key, set a new private_key_wo and bump the version
  // (also bump it when private_key_id changes)
  private_key_wo         = var.bigquery_metadata_private_key
  private_key_wo_version = 1

  private_key_id              = "my-private-key-id"
  client_email                = "my-service-account@my-project.iam.gserviceaccount.com"
  client_id                   = "123456789012345678901"
  auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
  token_uri                   = "https://oauth2.googleapis.com/token"
  auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
  client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/my-service-account%40my-project.iam.gserviceaccount.com"
}

// Service account key read from the JSON key file, with the classic sensitive attribute (stored in state)
locals {
  bigquery_key = jsondecode(file("service-account.json"))
}

resource "dbtcloud_bigquery_platform_metadata_credential" "example_from_file" {
  connection_id = dbtcloud_global_connection.bigquery.id

  cost_insights_enabled = true

  private_key                 = local.bigquery_key.private_key
  private_key_id              = local.bigquery_key.private_key_id
  client_email                = local.bigquery_key.client_email
  client_id                   = local.bigquery_key.client_id
  auth_uri                    = local.bigquery_key.auth_uri
  token_uri                   = local.bigquery_key.token_uri
  auth_provider_x509_cert_url = local.bigquery_key.auth_provider_x509_cert_url
  client_x509_cert_url        = local.bigquery_key.client_x509_cert_url
}

// One credential per connection, for example across many projects
resource "dbtcloud_bigquery_platform_metadata_credential" "per_connection" {
  for_each = dbtcloud_global_connection.bigquery_per_project

  connection_id         = each.value.id
  cost_insights_enabled = true

  private_key                 = local.bigquery_key.private_key
  private_key_id              = local.bigquery_key.private_key_id
  client_email                = local.bigquery_key.client_email
  client_id                   = local.bigquery_key.client_id
  auth_uri                    = local.bigquery_key.auth_uri
  token_uri                   = local.bigquery_key.token_uri
  auth_provider_x509_cert_url = local.bigquery_key.auth_provider_x509_cert_url
  client_x509_cert_url        = local.bigquery_key.client_x509_cert_url
}

// Workload Identity Federation. Requires a bigquery_v1 connection (use_latest_adapter = true).
// service_account_impersonation_url is filled in from the connection's impersonate_service_account
// when it is not set.
resource "dbtcloud_bigquery_platform_metadata_credential" "example_wif" {
  connection_id = dbtcloud_global_connection.bigquery_v1.id

  cost_insights_enabled = true

  auth_type                   = "external-oauth-wif"
  workload_pool_provider_path = "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/my-pool/providers/my-provider"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `connection_id` (Number) The ID of the global connection this credential is associated with. Cannot be changed after creation.

### Optional

- `auth_provider_x509_cert_url` (String) The `auth_provider_x509_cert_url` field of the service account JSON key. Required when `auth_type` is `service-account-json`.
- `auth_type` (String) The authentication type: `service-account-json` (default) or `external-oauth-wif`. `external-oauth-wif` requires a `bigquery_v1` connection. Changing it updates the credential in place.
- `auth_uri` (String) The `auth_uri` field of the service account JSON key. Required when `auth_type` is `service-account-json`.
- `catalog_ingestion_enabled` (Boolean) Whether catalog ingestion is enabled for this credential. When enabled, dbt Cloud will ingest metadata about tables, views, and other objects from your data warehouse.
- `client_email` (String) The `client_email` field of the service account JSON key. Required when `auth_type` is `service-account-json`.
- `client_id` (String) The `client_id` field of the service account JSON key. Required when `auth_type` is `service-account-json`.
- `client_x509_cert_url` (String) The `client_x509_cert_url` field of the service account JSON key. Required when `auth_type` is `service-account-json`.
- `cost_insights_enabled` (Boolean) Whether cost insights is enabled for this credential.
- `cost_optimization_enabled` (Boolean) Whether cost optimization data collection is enabled for this credential.
- `private_key` (String, Sensitive) The `private_key` field of the service account JSON key. Consider using `private_key_wo` instead, which is not stored in state. Exactly one of `private_key` and `private_key_wo` is required when `auth_type` is `service-account-json`.
- `private_key_id` (String) The `private_key_id` field of the service account JSON key. Required when `auth_type` is `service-account-json`.
- `private_key_wo` (String, Sensitive) Write-only alternative to `private_key`. The value is not stored in state. Requires `private_key_wo_version`.
- `private_key_wo_version` (Number) Version number for `private_key_wo`. Increment it to send a new `private_key_wo`, and whenever `private_key_wo` or `private_key_id` changes. To switch to the plain `private_key`, remove both `private_key_wo` and `private_key_wo_version`; the next apply sends the key again.
- `service_account_impersonation_url` (String) The service account impersonation URL, used when `auth_type` is `external-oauth-wif`. When it is not set and the connection has `impersonate_service_account`, dbt Cloud fills it in. Once set, removing it from the configuration does not clear it.
- `token_uri` (String) The `token_uri` field of the service account JSON key. Required when `auth_type` is `service-account-json`.
- `workload_pool_provider_path` (String) The full resource path of the workload identity pool provider. Required when `auth_type` is `external-oauth-wif`.

### Read-Only

- `adapter_version` (String) The adapter version derived from the connection (e.g., 'snowflake_v0', 'databricks_v0'). This is read-only and determined by the connection.
- `credential_id` (Number) The ID of the platform metadata credential.
- `id` (String) The unique identifier for this resource (account_id:credential_id).

## Import

Import is supported using the following syntax:

```shell
# The private key is never read back from dbt Cloud. After importing a credential that uses
# a service account key, the first apply sends the private key again. This is expected.

# using import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_bigquery_platform_metadata_credential.my_credential
  id = "credential_id"
}

import {
  to = dbtcloud_bigquery_platform_metadata_credential.my_credential
  id = "1234"
}

# using the older import command
terraform import dbtcloud_bigquery_platform_metadata_credential.my_credential "credential_id"
terraform import dbtcloud_bigquery_platform_metadata_credential.my_credential 1234
```
