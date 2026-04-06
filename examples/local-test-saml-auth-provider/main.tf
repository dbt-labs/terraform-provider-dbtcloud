terraform {
  required_version = ">= 1.11"
  required_providers {
    dbtcloud = {
      source  = "dbt-labs/dbtcloud"
      version = "~> 1.0"
    }
  }
}

provider "dbtcloud" {
}

locals {
  # Public test certificate from mocksaml.com — not sensitive.
  cert = <<-EOT
    xxx
  EOT
}

# ─────────────────────────────────────────────────────────────────────────────
# Uncomment ONE case at a time — only one auth provider may exist per account.
# Run `terraform destroy` between cases.
# ─────────────────────────────────────────────────────────────────────────────


# ── Case 1: SAML — write-only cert (not stored in state) ─────────────────────
# Verify: login_url and slug are computed, cert_expiry_date is set,
#         cert value is absent from terraform.tfstate.
# Cert rotation: bump cert_wo_version to trigger an update without destroy.

# resource "dbtcloud_auth_provider" "case1" {
#   type            = "saml"
#   entity_id       = "https://saml.example.com/entityid"
#   sso_url         = "https://mocksaml.com/api/saml/sso"
#   cert_wo         = local.cert
#   cert_wo_version = 1
#   allow_password_backdoor = true
# }
# output "case1" { value = dbtcloud_auth_provider.case1 }


# ── Case 2: SAML — state-stored cert, optional fields ────────────────────────
# Verify: cert appears as (sensitive) in state, sign_request and attribute_map
#         round-trip correctly, allow_password_backdoor=false is accepted.

# resource "dbtcloud_auth_provider" "case2" {
#   type      = "saml"
#   entity_id = "https://saml.example.com/entityid"
#   sso_url   = "https://mocksaml.com/api/saml/sso"
#   cert      = local.cert
#   sign_request  = true
#   attribute_map = jsonencode({
#     email      = "nameID"
#     first_name = "firstName"
#     last_name  = "lastName"
#   })
#   allow_password_backdoor = false
# }
# output "case2" { value = dbtcloud_auth_provider.case2 }


# ── Case 3: Okta ──────────────────────────────────────────────────────────────
# Identical to SAML on the API side — just validates the type is accepted.
# Verify: creates successfully, cert_expiry_date is populated.

# resource "dbtcloud_auth_provider" "case3" {
#   type            = "okta"
#   entity_id       = "https://saml.example.com/entityid"
#   sso_url         = "https://mocksaml.com/api/saml/sso"
#   cert_wo         = local.cert
#   cert_wo_version = 1
# }
# output "case3" { value = dbtcloud_auth_provider.case3, sensitive = true }


# ── Case 4: Azure AD — single tenant ─────────────────────────────────────────
# Verify: creates successfully, domain and include_indirect_groups round-trip,
#         client_secret does not appear in state in plaintext.
# Credentials: set via env vars or a local .tfvars (not committed).

# resource "dbtcloud_auth_provider" "case4" {
#   type                    = "azure_single_tenant"
#   client_id               = var.azure_client_id
#   client_secret_wo        = var.azure_client_secret
#   client_secret_wo_version = 1
#   tenant_id               = var.azure_tenant_id
#   include_indirect_groups = true
#   max_groups_to_retrieve  = 500
# }
# output "case4" { value = dbtcloud_auth_provider.case4, sensitive = true }


# ── Case 5: Azure AD — multi tenant ──────────────────────────────────────────
# Same credentials, no tenant_id required.
# Verify: creates without tenant_id, plan is clean on re-apply.

# resource "dbtcloud_auth_provider" "case5" {
#   type                    = "azure_multi_tenant"
#   client_id               = var.azure_client_id
#   client_secret_wo        = var.azure_client_secret
#   client_secret_wo_version = 1
# }
# output "case5" { value = dbtcloud_auth_provider.case5, sensitive = true }


# ── Case 6: Google Workspace ──────────────────────────────────────────────────
# Verify: creates successfully, gsuite_admin_id and domain round-trip correctly,
#         admin_refresh_token appears as (sensitive) in state.
# Credentials: set via env vars or a local .tfvars (not committed).

# resource "dbtcloud_auth_provider" "case6" {
#   type                = "gsuite"
#   client_id           = var.gsuite_client_id
#   client_secret       = var.gsuite_client_secret
#   admin_refresh_token = var.gsuite_refresh_token
#   domain              = var.gsuite_domain
#   gsuite_admin_id     = var.gsuite_admin_email
# }
# output "case6" { value = dbtcloud_auth_provider.case6, sensitive = true }


# ── Case 7: Import ────────────────────────────────────────────────────────────
# Run after any successful apply above. Get the ID from state:
#   terraform show | grep '"id"'
# Then:
#   terraform destroy
#   # uncomment the resource below, then:
#   terraform import dbtcloud_auth_provider.case7 <id>
#   terraform plan   # expect: no changes except secret fields

# resource "dbtcloud_auth_provider" "case7" {
#   type            = "saml"
#   entity_id       = "https://saml.example.com/entityid"
#   sso_url         = "https://mocksaml.com/api/saml/sso"
#   cert_wo         = local.cert
#   cert_wo_version = 1
# }


# ── Case 8: Destroy → recreate different type ─────────────────────────────────
# Verifies no 409 conflict after destroy. Steps:
#   1. Apply case 1 (saml), then destroy.
#   2. Uncomment this block and apply — should succeed with no conflict.
# Verify: new resource created, different auto-generated slug.

# resource "dbtcloud_auth_provider" "case8" {
#   type            = "okta"
#   entity_id       = "https://saml.example.com/entityid"
#   sso_url         = "https://mocksaml.com/api/saml/sso"
#   cert_wo         = local.cert
#   cert_wo_version = 1
# }
# output "case8" { value = dbtcloud_auth_provider.case8 }


# ── Case 9: Validation errors (all should fail at `terraform plan`) ───────────
# Uncomment one block at a time and run `terraform plan` — expect a clear error.

# 9a: missing entity_id
# resource "dbtcloud_auth_provider" "case9a" {
#   type    = "saml"
#   sso_url = "https://mocksaml.com/api/saml/sso"
#   cert_wo = local.cert
# }

# 9b: missing sso_url
# resource "dbtcloud_auth_provider" "case9b" {
#   type      = "saml"
#   entity_id = "https://saml.example.com/entityid"
#   cert_wo   = local.cert
# }

# 9c: missing cert entirely
# resource "dbtcloud_auth_provider" "case9c" {
#   type      = "saml"
#   entity_id = "https://saml.example.com/entityid"
#   sso_url   = "https://mocksaml.com/api/saml/sso"
# }

# 9d: cert and cert_wo both set (conflict)
# resource "dbtcloud_auth_provider" "case9d" {
#   type      = "saml"
#   entity_id = "https://saml.example.com/entityid"
#   sso_url   = "https://mocksaml.com/api/saml/sso"
#   cert      = local.cert
#   cert_wo   = local.cert
# }

# 9e: client_secret and client_secret_wo both set (conflict)
# resource "dbtcloud_auth_provider" "case9e" {
#   type             = "azure_single_tenant"
#   client_id        = "some-id"
#   client_secret    = "secret"
#   client_secret_wo = "secret"
#   tenant_id        = "some-tenant"
# }

# 9f: azure_single_tenant missing tenant_id
# resource "dbtcloud_auth_provider" "case9f" {
#   type             = "azure_single_tenant"
#   client_id        = "some-id"
#   client_secret_wo = "secret"
# }
