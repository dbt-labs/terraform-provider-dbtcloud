terraform {
  required_providers {
    dbt = {
      source  = "GtheSheep/dbt-cloud"
      version = "0.0.39"
    }
  }
}

provider "dbt" {
  account_id = <ACCOUNT_ID>
  token      = "<TOKEN>>"
}
