---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dbtcloud_postgres_credential Data Source - dbtcloud"
subcategory: ""
description: |-
  Postgres credential data source.
---

# dbtcloud_postgres_credential (Data Source)

Postgres credential data source.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `credential_id` (Number) Credential ID
- `project_id` (Number) Project ID

### Read-Only

- `default_schema` (String) Default schema name
- `id` (String) The ID of this data source. Contains the project ID and the credential ID.
- `is_active` (Boolean) Whether the Postgres credential is active
- `num_threads` (Number) Number of threads to use
- `username` (String) Username for Postgres
