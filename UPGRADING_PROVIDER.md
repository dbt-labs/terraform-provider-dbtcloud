# Upgrading from the GtheSheep Provider

As of 6/14/2023 the provider has been transferred from the dbt community member Gary James [[GtheSheep](https://github.com/GtheSheep)] to dbt-labs.

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

To change the version of the provider, please also run the following command

```sh
terraform init -upgrade
```

>**Note**:  0.1.12 is the first version published after the transfer. For earlier versions, please continue using the GtheSheep/dbt-cloud source

>**Note**:  Once you verify that the new provider is working fine with a `terraform plan`, you can also change the version to the latest 0.2.x and run another `terraform init -upgrade` to get the latest resource parameters and bug fixes

---

# Upgrading from the `dbt_cloud_xxx` resources to the `dbtcloud_xxx` ones

With version 0.2, resources and data sources are both available as `dbt_cloud_xxx` (legacy) and `dbtcloud_xxx` (preferred, following the Terraform convention).

- `dbt_cloud_xxx` is kept in 0.2 for backward compatibility, but will be removed from version 0.3 onwards. Consider starting new projects with the `dbtcloud_xxx` naming convention
- `dbtcloud_xxx` follows the Terraform naming convention and is the long term convention for the dbt Cloud configuration
- a single Terraform project can't use this dbt Cloud Terraform provider with both `dbt_cloud_xxx` and `dbtcloud_xxx` resources at the same time

More details:

| `dbt_cloud_xxx` resources  | `dbtcloud_xxx` resources |
| ------------- | ------------- |
| in v0.2.x, the resources which existed previously in the GtheSheep version are still available  | every resource that existed as `dbt_cloud_xxx` also exists as `dbtcloud_xxx` in 0.2.x   |
| when using newer version of the provider, those resources will leverage the bug fixes and new parameters for the existing resources | some new resources have been introduced since 0.2.x and only exist as `dbtcloud_xxx`. <br/>As mentioned above, using those new resources will require moving all the previous ones from `dbt_cloud_xxx` to `dbtcloud_xxx` (see below) |

## Handling the move from `dbt_cloud_xxx` (legacy) to `dbtcloud_xxx`

As those are different resources, it is not possible to move existing resources using the `terraform state mv` command.

The options are:

- keep existing projects with `dbt_cloud_xxx` resources, and create new project with `dbtcloud_xxx` resources. Older projects won't be able to use new resource types.
- or update the state file manually to change the resource names (this should work but it is possible to corrupt the state, be careful and keep a backup)

Here is an example of how to update the state file:

1. perform a `terraform apply` to make sure that no changes are required between the current config and dbt Cloud
1. edit the resource configuration files, changing resources from `dbt_cloud_xxx` to `dbtcloud_xxx`
   - the resource types need to be updated where the resources are defined but also wherever we are referring to other resource (e.g. in `name = dbt_cloud_project.name`)
1. edit `required_providers { dbt  = {` and `provider "dbt"` to `required_providers { dbtcloud  = {` and `provider "dbtcloud"`
1. pull the remote state with `terraform state pull > remote_state.tfstate` and keep a back up of the file
1. edit the state file to change the resource types from `dbt_cloud_xxx` to `dbtcloud_xxx`
   - update where there is a `"type": "dbtcloud_xxx"`
   - and where there is a `"dependencies": ["dbtcloud_xxx.abc"]`
1. push the state back with `terraform state push remote_state.tfstate`
   - you might need to do a `terraform state push -force remote_state.tfstate`
1. perform a `terraform init -upgrade` to update the terraform provider
1. perform a `terraform plan` to check that no change is required, you can then delete the backup of the state
