// SAML — write-only cert (recommended, not stored in state)
//
// Requires Terraform >= 1.11 for write-only attribute support.
// Bump cert_wo_version to rotate the cert without recreating the resource.

variable "saml_cert" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_auth_provider" "saml" {
  type      = "saml"
  entity_id = "https://your-idp.example.com/metadata"
  sso_url   = "https://your-idp.example.com/sso/saml"

  cert_wo         = var.saml_cert
  cert_wo_version = 1
}

output "login_url" {
  description = "SSO login URL to share with users."
  value       = dbtcloud_auth_provider.saml.login_url
}


// SAML — all optional fields

resource "dbtcloud_auth_provider" "saml_full" {
  type      = "saml"
  entity_id = "https://your-idp.example.com/metadata"
  sso_url   = "https://your-idp.example.com/sso/saml"
  cert      = file("idp-cert.pem")

  sign_request  = true
  attribute_map = jsonencode({
    email      = "nameID"
    first_name = "firstName"
    last_name  = "lastName"
  })

  allow_password_backdoor = false
}


// Okta (identical to SAML, different type value)

resource "dbtcloud_auth_provider" "okta" {
  type      = "okta"
  entity_id = "http://www.okta.com/<okta_app_id>"
  sso_url   = "https://<your-org>.okta.com/app/<app_path>/sso/saml"

  cert_wo         = var.saml_cert
  cert_wo_version = 1
}


// Azure AD — single tenant

variable "azure_client_secret" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_auth_provider" "azure_single_tenant" {
  type      = "azure_single_tenant"
  client_id = "00000000-0000-0000-0000-000000000000"
  tenant_id = "11111111-1111-1111-1111-111111111111"

  client_secret_wo         = var.azure_client_secret
  client_secret_wo_version = 1

  domain                  = "acme.com"
  include_indirect_groups = true
  max_groups_to_retrieve  = 500
}


// Azure AD — multi tenant (no tenant_id required)

resource "dbtcloud_auth_provider" "azure_multi_tenant" {
  type      = "azure_multi_tenant"
  client_id = "00000000-0000-0000-0000-000000000000"

  client_secret_wo         = var.azure_client_secret
  client_secret_wo_version = 1
}


// Azure Active Directory

resource "dbtcloud_auth_provider" "azure_active_directory" {
  type      = "azure_active_directory"
  client_id = "00000000-0000-0000-0000-000000000000"
  tenant_id = "11111111-1111-1111-1111-111111111111"

  client_secret_wo         = var.azure_client_secret
  client_secret_wo_version = 1

  domain = "acme.com"
}


// Google Workspace

variable "gsuite_client_secret" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_auth_provider" "gsuite" {
  type      = "gsuite"
  client_id = "000000000000-xxxx.apps.googleusercontent.com"

  client_secret_wo         = var.gsuite_client_secret
  client_secret_wo_version = 1

  admin_refresh_token = "<oauth-refresh-token>"
  domain              = "acme.com"
  gsuite_admin_id     = "admin@acme.com"
}
