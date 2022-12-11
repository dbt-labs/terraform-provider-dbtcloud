terraform {
  required_providers {
    dbt = {
      source  = "GtheSheep/dbt_cloud"
      version = "0.1.0"
    }
  }
}

provider "dbt" {
  account_id = var.dbt_cloud_account_id
  token      = var.dbt_cloud_token
  host_url   = "https://cloud.getdbt.com/api"
}
