terraform {
  required_providers {
    dbtcloud = {
      source = "dbt-labs/dbtcloud"
    }
  }
}

provider "dbtcloud" {}

variable "snowflake_password" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_project" "test" {
  name = "write-only-test"
}

resource "dbtcloud_snowflake_credential" "test" {
  project_id          = dbtcloud_project.test.id
  auth_type           = "password"
  database            = "MY_DB"
  role                = "MY_ROLE"
  warehouse           = "MY_WH"
  schema              = "MY_SCHEMA"
  user                = "MY_USER"
  password_wo         = var.snowflake_password
  password_wo_version = 1
  num_threads         = 3
}
