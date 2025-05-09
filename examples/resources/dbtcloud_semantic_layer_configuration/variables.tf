// terraform apply -var="project_id=12345" -var="environment_id=67890"

variable "project_id" {
  description = "The ID of the dbt Cloud project"
  type        = number
}

variable "environment_id" {
  description = "The ID of the dbt Cloud environment (must have at least one successful run)"
  type        = number
} 