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
