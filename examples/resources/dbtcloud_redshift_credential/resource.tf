// Using the classic sensitive attribute (stored in state)
resource "dbtcloud_redshift_credential" "redshift" {
  num_threads    = 16
  project_id     = dbtcloud_project.test_project.id
  default_schema = "my_schema"
  // example of optional fields
  username       = "my_username"
  password       = "my_sensitive_password"
  is_active      = true
}

// Using write-only attributes (not stored in state, requires Terraform >= 1.11)
//
// The password_wo value is never persisted in the Terraform state file.
// Use password_wo_version to trigger an update when the password changes.
variable "redshift_password" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_redshift_credential" "redshift_wo" {
  num_threads         = 16
  project_id          = dbtcloud_project.test_project.id
  default_schema      = "my_schema"
  username            = "my_username"
  password_wo         = var.redshift_password
  password_wo_version = 1
  is_active           = true
}
