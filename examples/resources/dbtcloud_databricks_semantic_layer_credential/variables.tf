variable "project_id" {
  description = "The ID of the dbt Cloud project"
  type        = number
}

variable "catalog" {
  description = "The catalog where to create models (only for the databricks adapter)"
  type        = string
}

variable "token" {
  description = "Token for Databricks user"
  type        = string
}

variable "semantic_layer_credential" {
  description = "This field indicates that the credential is used as part of the Semantic Layer configuration. It is used to create a Databricks credential for the Semantic Layer."
  type        = bool
}
