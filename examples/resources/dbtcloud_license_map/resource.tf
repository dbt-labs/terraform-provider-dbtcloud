# Developer license group mapping
resource "dbtcloud_license_map" "dev_license_map" {
  license_type = "developer"
  sso_license_mapping_groups = ["DEV-SSO-GROUP"]
}

# Read-only license mapping
resource "dbtcloud_license_map" "read_only_license_map" {
  license_type = "read-only"
  sso_license_mapping_groups = ["READ-ONLY-SSO-GROUP"]
}

# IT license mapping
resource "dbtcloud_license_map" "it_license_map" {
  license_type = "it"
  sso_license_mapping_groups = ["IT-SSO-GROUP"]
}
