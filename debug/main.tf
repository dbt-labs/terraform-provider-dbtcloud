terraform {
  required_providers {
    dbt = {
      source  = "GtheSheep/dbt-cloud"
      version = "0.0.67"
    }
  }
}

variable "dbt_cloud_project_id" {
  type        = number
  description = "DBT Cloud Project ID"
}

data "dbt_cloud_project" "dbt_cloud_project" {
  project_id = var.dbt_cloud_project_id
}
