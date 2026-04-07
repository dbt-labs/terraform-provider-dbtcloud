# Use a native OpenAI API key.
# Recommended: write-only (Terraform 1.11+) — the key is never stored in state.
resource "dbtcloud_openai_integration" "openai" {
  key_type             = "openai"
  key_value_wo         = var.openai_api_key
  key_value_wo_version = 1
}

# For older Terraform versions, use key_value instead — stored as a sensitive value in state.
# resource "dbtcloud_openai_integration" "openai" {
#   key_type  = "openai"
#   key_value = var.openai_api_key
# }

# Use an Azure OpenAI deployment.
resource "dbtcloud_openai_integration" "azure" {
  key_type              = "azure_openai"
  key_value_wo          = var.azure_openai_api_key
  key_value_wo_version  = 1
  azure_endpoint        = "https://my-deployment.openai.azure.com/"
  azure_deployment_name = "gpt-4o"
  azure_api_version     = "2024-02-01"
}

# To revert to the dbt Labs-managed key, remove this resource from your config
# and run `terraform apply`, or run `terraform destroy`. Deleting the record is
# all that is needed — dbt Cloud automatically falls back to the managed key
# when no integration exists.
