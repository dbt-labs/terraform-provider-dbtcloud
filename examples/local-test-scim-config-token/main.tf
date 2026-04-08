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
  token      = "xxxx"
  host_url   = "https://dbt.com/api"
}

# ── Case 1: SCIM global config ────────────────────────────────────────────────
# Start: enabled=true, manual_updates_allowed=false, scim_controlled_license_type=false
# Then toggle to test updates:
#   Step 2 — manual_updates_allowed=true
#   Step 3 — scim_controlled_license_type=true
#   Step 4 — enabled=false (disables SCIM entirely)
resource "dbtcloud_scim_config" "main" {
  enabled                      = true
  manual_updates_allowed       = true
  scim_controlled_license_type = true
}

# ── Case 2: create a SCIM token ───────────────────────────────────────────────
resource "dbtcloud_scim_config_token" "test" {
  name = "tf-test-scim"
}

output "scim_token" {
  value     = dbtcloud_scim_config_token.test.token_string
  sensitive = true
}

# ── Case 3: read back current config (no-op plan) ─────────────────────────────
# After applying Case 1, comment it out and uncomment this to verify no drift.
# resource "dbtcloud_scim_config" "main" {
#   enabled                      = true
#   manual_updates_allowed       = false
#   scim_controlled_license_type = false
# }

# ── Case 4: enable manual updates ─────────────────────────────────────────────
# Uncomment and re-apply to test updating manual_updates_allowed=true.
# resource "dbtcloud_scim_config" "main" {
#   enabled                      = true
#   manual_updates_allowed       = true
#   scim_controlled_license_type = false
# }

# ── Case 5: enable scim_controlled_license_type ───────────────────────────────
# Uncomment and re-apply to test enabling license type control via SCIM.
# resource "dbtcloud_scim_config" "main" {
#   enabled                      = true
#   manual_updates_allowed       = true
#   scim_controlled_license_type = true
# }

# ── Case 6: disable SCIM entirely ─────────────────────────────────────────────
# Uncomment and re-apply to verify enabled=false is accepted and reflected.
# resource "dbtcloud_scim_config" "main" {
#   enabled                      = false
#   manual_updates_allowed       = false
#   scim_controlled_license_type = false
# }
