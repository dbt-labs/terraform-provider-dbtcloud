variable "project_id" {
  description = "The ID of the dbt Cloud project"
  type        = number
}

variable "username" {
  description = "The Snowflake database name"
  type        = string
}

variable "password" {
  description = "The password for the Redshift account"
  type        = string
}

variable "num_threads" {
  description = "Number of threads to use"
  type        = string
}

variable "default_schema" {
  description = "Default schema name"
  type        = string
}