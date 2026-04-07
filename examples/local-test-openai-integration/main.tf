terraform {
  required_providers {
    dbtcloud = {
      source  = "dbt-labs/dbtcloud"
      version = ">= 0.3"
    }
  }
}

provider "dbtcloud" {
  account_id = 1234
  token      = "xxx"
  host_url   = "https://dbt.com/api"
}

# ── Case 1: native OpenAI key (key_value, stored in state) ────────────────────
# resource "dbtcloud_openai_integration" "test" {
#   key_type  = "openai"
#   key_value = "sk-test-placeholder"
# }

# ── Case 2: rotate the key ────────────────────────────────────────────────────
# Uncomment and re-apply to test PATCH updating key_value.
# resource "dbtcloud_openai_integration" "test" {
#   key_type  = "openai"
#   key_value = "sk-test-rotated"
# }

# ── Case 4: azure_openai ──────────────────────────────────────────────────────
# Uncomment and re-apply to test Azure OpenAI integration.
# resource "dbtcloud_openai_integration" "test" {
#   key_type              = "azure_openai"
#   key_value             = "az-test-placeholder"
#   azure_endpoint        = "https://my-deployment.openai.azure.com/"
#   azure_deployment_name = "gpt-4o"
#   azure_api_version     = "2024-02-01"
# }

# ── Case 5: write-only key (Terraform 1.11+) ──────────────────────────────────
# Uncomment and re-apply to test key_value_wo (not stored in state).
# resource "dbtcloud_openai_integration" "test" {
#   key_type             = "openai"
#   key_value_wo         = "sk-wo-placeholder"
#   key_value_wo_version = 1
# }
