---
page_title: "Migrating from deprecated resources"
subcategory: ""
---

# Migrating from deprecated resources

Starting with release v1.0.0, the following resources have been removed:
- `dbtcloud_bigquery_connection`
- `dbtcloud_connection` (enveloped connections for databricks, redshift and snowflake)
- `dbtcloud_fabric_connection`
- `dbtcloud_project_connection`
All of these resources, excluding `dbtcloud_project_connection` which became redundant, are to be replaced by the `dbtcloud_global_connection` resource.

Bellow will be two examples on how the `dbtcloud_global_connection` differs from the two deprecated variants for warehouse connections.

```terraform
# Example for dbtcloud_connection, the abstraced connection resource variant

resource "dbtcloud_connection" "deprecated_redshift" {
  project_id = dbtcloud_project.dbt_project.id
  type       = "redshift"
  name       = "My Redshift Warehouse"
  database   = "my-database"
  port       = 5439
  host_name  = "my-redshift-hostname"
}

resource "dbtcloud_global_connection" "new_redshift" {
  name = "My Redshift connection"
  redshift = {
    hostname = "my-redshift-connection.com"
    port     = 5432
    dbname = "my_database"
  }
}

# Example for dbtcloud_fabric_connection, the dedicated connection resource variant

resource "dbtcloud_fabric_connection" "deprecated_fabric" {
  project_id    = dbtcloud_project.dbt_project.id
  name          = "Connection Name"
  server        = "my-server"
  database      = "my-database"
  port          = 1234
  login_timeout = 30
}

resource "dbtcloud_global_connection" "new_fabric" {
  name = "My Fabric connection"
  fabric = {
    server   = "my-fabric-server.com"
    database = "mydb"
    port          = 1234
    retries       = 3
    login_timeout = 60
    query_timeout = 3600
  }
}
```

This guide shows how to migrate from a resource which has been deprecated or renamed to its replacement.
It's possible to migrate between the resources by updating your Terraform Configuration, removing the old state, and the importing the new resource in config.

In this guide, we'll highlight the migration from `dbtcloud_connection` to `dbtcloud_global_connection`. This is also applicable to any other resources that need to be migrated.

Assuming we have the following Terraform Configuration:
```terraform
resource "dbtcloud_project" "dbt_project" {
    # ...
}

resource "dbtcloud_connection" "redshift" {
  project_id = dbtcloud_project.dbt_project.id
  type       = "redshift"
  name       = "My Redshift Warehouse"
  database   = "my-database"
  port       = 5439
  host_name  = "my-redshift-hostname"
}
```

We can update the Terraform Configuration to use the new resource by updating the resource to the new `dbtcloud_global_connection` schema:
```terraform
resource "dbtcloud_project" "dbt_project" {
    # ...
}

resource "dbtcloud_global_connection" "redshift" {
  name = "My Redshift connection"
  redshift = {
    hostname = "my-redshift-connection.com"
    port     = 5432
    dbname = "my_database"
  }
}
```

As the Terraform Configuration has been updated - we now need to update the State. We can view the items Terraform is tracking in its statefile using the `terraform state list` command, for example:
```bash
$ terraform state list
dbtcloud_connection.redshift
dbtcloud_project.dbt_project
```

In order to migrate from the old resource to the new resource we need to first remove the old resource from the state - and subsequently use Terraform's [import functionality](https://www.terraform.io/docs/import/index.html) to migrate to the new resource.
To import a resource in Terraform we first require its Resource ID - we can obtain this from the command-line via:

```shell
$ echo dbtcloud_connection.redshift.id | terraform console
```

Next we can remove the existing resource using `terraform state rm` - for example:

```shell
$ terraform state rm dbtcloud_connection.redshift
```

Now that the old resource has been removed from Terraform's Statefile we can now Import it into the Statefile as the new resource by running:

```text
terraform import dbtcloud_global_connection.redshift [resourceid]
```

Once this has been done, running terraform plan should show no changes since at this point, you've switched over to using the new resource and should be able to continue using Terraform as normal.