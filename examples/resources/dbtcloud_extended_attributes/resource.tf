# extended_attributes can be set as a raw JSON string or encoded with Terraform's `jsonencode()` function
# we recommend using `jsonencode()` to avoid Terraform reporting changes due to whitespaces or keys ordering
resource "dbtcloud_extended_attributes" "my_attributes" {
  extended_attributes = jsonencode(
    {
      type      = "databricks"
      catalog   = "dbt_catalog"
      http_path = "/sql/your/http/path"
      my_nested_field = {
        subfield = "my_value"
      }
    }
  )
  project_id = var.dbt_project.id
}

resource "dbtcloud_environment" "issue_depl" {
  dbt_version            = "latest"
  name                   = "My environment"
  project_id             = var.dbt_project.id
  type                   = "deployment"
  use_custom_branch      = false
  credential_id          = var.dbt_credential_id
  deployment_type        = "production"
  extended_attributes_id = dbtcloud_extended_attributes.my_attributes.extended_attributes_id
}