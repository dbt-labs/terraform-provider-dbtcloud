terraform {
  required_providers {
    dbtcloud = {
      source  = "dbt-labs/dbtcloud"
      version = "~> 1.0"
    }
  }
}

provider "dbtcloud" {
  account_id = var.dbt_cloud_account_id
  token      = var.dbt_cloud_token
  host_url   = "https://cloud.getdbt.com/api"
  timeout_seconds = 60
  max_retries = 10
  retry_interval_seconds = 10
  disable_retry = false
  skip_credentials_validation = false
  retriable_status_codes = ["429", "500", "502", "503", "504"]
}
