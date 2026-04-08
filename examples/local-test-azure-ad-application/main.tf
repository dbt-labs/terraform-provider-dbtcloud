terraform {
  required_providers {
    dbtcloud = {
      source  = "dbt-labs/dbtcloud"
      version = ">= 0.3"
    }
  }
}

provider "dbtcloud" {
  account_id = 1234       # replace with your account ID
  token      = "xxxx"     # replace with your API token
  host_url   = "https://dbt.com/api"  # replace with your dbt Cloud host URL if different
}

# ── Case 1: Create with service_user auth (default) ──────────────────────────
# Apply this first. Verify in the dbt Cloud UI under Account Settings > Integrations
# that the Azure AD application appears with the correct org name.
resource "dbtcloud_azure_ad_application" "test" {
  organization_name = "xxxx"
  client_id         = "xxxx"
  client_secret     = "xxxx"
  tenant_id         = "xxxx"

  azure_service_authentication_method = "service_user"
}

output "azure_ad_application_id" {
  value = dbtcloud_azure_ad_application.test.id
}

output "oauth_redirect_uri_domain" {
  description = "The OAuth redirect URI domain set by the API (computed)"
  value       = dbtcloud_azure_ad_application.test.oauth_redirect_uri_domain
}

# ── Case 2: Switch auth method to service_principal ───────────────────────────
# Comment out Case 1 above and uncomment this block, then re-apply.
# Verify the auth method changes in the UI without recreating the resource.
# resource "dbtcloud_azure_ad_application" "test" {
#   organization_name = "xxxx"
#   client_id         = "xxxx"
#   client_secret     = "xxxx"
#   tenant_id         = "xxxx"
#
#   azure_service_authentication_method = "service_principal"
# }

# ── Case 3: Rotate the client_secret ─────────────────────────────────────────
# Change client_secret to a new value and re-apply.
# The API requires all four fields (org, client_id, secret, tenant_id) on every
# update — they are stored in state as sensitive so they can be resent.
# resource "dbtcloud_azure_ad_application" "test" {
#   organization_name = "my-azure-devops-org"
#   client_id         = "00000000-0000-0000-0000-000000000000"
#   client_secret     = "my-NEW-client-secret"    # rotated
#   tenant_id         = "00000000-0000-0000-0000-000000000001"
#
#   azure_service_authentication_method = "service_user"
# }

# ── Case 4: Destroy ────────────────────────────────────────────────────────────
# Run: terraform destroy
# Verify the Azure AD application is removed from Account Settings > Integrations.
