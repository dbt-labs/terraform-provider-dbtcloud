# Upgrading from the GtheSheep Provider

As of (6/14/2023) the provider has been transferred from the dbt community member Gary James [[GtheSheep](https://github.com/GtheSheep)] to dbt-labs.

To upgrade from the community provider to the dbt-labs one, please run the following command:

```shell
terraform state replace-provider GtheSheep/dbt-cloud dbt-labs/dbtcloud
```

You should also update your lock file / Terraform provider version pinning. From the deprecated source:

```hcl
# deprecated source
terraform {
  required_providers {
    dbt = {
      source  = "GtheSheep/dbt-cloud"
      version = "0.1.11"
    }
  }
}
```

To new source:

```hcl
# new source
terraform {
  required_providers {
    dbt = {
      source  = "dbt-labs/dbtcloud"
      version = "0.1.12"
    }
  }
}
```

To change version of the provider please also run the following command

```sh
terraform init -upgrade
```

>**Note**:  0.1.12 is the first version published after the transfer. For earlier versions, please continue using the GtheSheep/dbt-cloud source


# Upgrading from the `dbt_cloud_xxx` resources to the `dbtcloud_xxx` ones

With version 0.2, resources and data sources are both available as `dbt_cloud_xxx` (legacy) and `dbtcloud_xxx` (preferred, following the Terraform convention).

- `dbt_cloud_xxx` is kept in 0.2 for backward compatibility, but will be removed from version 0.3 onwards. Consider starting new projects with the `dbtcloud_xxx` naming convention
- `dbtcloud_xxx` follows the Terraform naming convention and is the long term convention for the dbt Cloud configuration

## Handling the move from `dbt_cloud_xxx` (legacy) to `dbtcloud_xxx`

As those are different resources, it is not possible to move existing resources using the `terraform state mv` command.

The options are:

- keep existing projects with `dbt_cloud_xxx` resources, and create new ones with `dbtcloud_xxx`
- or update the state file manually to change the resource names (this should work but it is possible to corrupt the state, be careful and keep a backup)
  1. perform a `terraform apply` to apply the changes required to dbt Cloud
  1. edit the resource configuration files changing resources from `dbt_cloud_xxx` to `dbtcloud_xxx`
  1. edit `required_providers { dbt  = {` and `provider "dbt"` to `required_providers { dbtcloud  = {` and `provider "dbtcloud"`
  1. pull the remote state with `terraform state pull > remote_state.tfstate` and keep a back up of the file
  1. edit the state file to change the resource types from `dbt_cloud_xxx` to `dbtcloud_xxx`
  1. push the state back with `terraform state push remote_state.tfstate`
  1. perform a `terraform init -upgrade` to update the terraform provider
  1. perform a `terraform plan` to check that no change is required, you can then delete the backup of the state