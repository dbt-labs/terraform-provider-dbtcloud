terraform {
  required_providers {
    dbtcloud = {
      source  = "dbt-labs/dbtcloud"
      version = "= 1.1.0"
    }
  }
}

variable "project_id" {
  description = "The ID of the Google Cloud project"
  type        = string
}

variable "environment_id" {
  description = "The ID of the environment"
  type        = string
}

variable "client_slug" {
  description = "The client slug used for naming resources"
  type        = string
}

resource "dbtcloud_semantic_layer_configuration" "this" {
  project_id     = var.project_id
  environment_id = var.environment_id
}

resource "dbtcloud_redshift_semantic_layer_credential" "this" {
  configuration = {
    project_id      = var.project_id
    name            = "Semantic - Redshift - ${var.client_slug}"
    adapter_version = "redshift_v0"
  }

  credential = {
    project_id     = var.project_id
    is_active      = true
    default_schema = "${var.client_slug}_custom"
    username       = "dbtcloud_${var.client_slug}"
    password       = "abcdefghijklmnop"
    num_threads    = 0
  }
}