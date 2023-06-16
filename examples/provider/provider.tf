terraform {
  required_providers {
    dbt = {
      source  = "dbt-labs/dbtcloud"
      version = "0.1.12"
    }
  }
}

provider "dbt" {
  account_id = var.dbt_cloud_account_id
  token      = var.dbt_cloud_token
  host_url   = "https://cloud.getdbt.com/api"
}
