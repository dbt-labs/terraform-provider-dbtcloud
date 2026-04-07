# Enable SCIM with strict IdP control — no manual updates, license type managed by SCIM.
# This is the recommended configuration when using an identity provider such as
# Okta or Azure AD as the single source of truth for user provisioning.
resource "dbtcloud_scim_config" "main" {
  enabled                      = true
  manual_updates_allowed       = false
  scim_controlled_license_type = true
}
