variable "project_id" {
  description = "The ID of the dbt Cloud project"
  type        = number
}

variable "username" {
  description = "The Postgres database name"
  type        = string
}

variable "password" {
  description = "The password for the Postgres account"
  type        = string
}