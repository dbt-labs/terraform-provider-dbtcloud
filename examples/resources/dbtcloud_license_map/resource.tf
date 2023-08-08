resource "dbtcloud_license_map" "test_license_map" {
  license_type = "developer"
  sso_license_mapping_groups = ["TEST-GROUP"]
}
