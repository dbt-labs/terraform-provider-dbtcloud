terraform {
  required_providers {
    dbt = {
      source  = "GtheSheep/dbt-cloud"
      version = "0.0.75"
    }
  }
}

provider "dbt" {
  account_id = var.dbt_cloud_account_id
  token      = var.dbt_cloud_token
}
