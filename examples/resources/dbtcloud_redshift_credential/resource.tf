resource "dbtcloud_redshift_credential" "redshift" {
  num_threads    = 16
  project_id     = dbtcloud_project.test_project.id
  default_schema = "my_schema"
  // example of optional fields
  username       = "my_username"
  password       = "my_sensitive_password"
  is_active      = true
}