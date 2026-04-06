// Using the classic sensitive attribute (stored in state)
resource "dbtcloud_teradata_credential" "example" {
  project_id = dbtcloud_project.example.id
  schema     = "your_schema"
  user       = "your_user"
  password   = "your_password"
}

// Using write-only attributes (not stored in state, requires Terraform >= 1.11)
//
// The password_wo value is never persisted in the Terraform state file.
// Use password_wo_version to trigger an update when the password changes.
variable "teradata_password" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_teradata_credential" "example_wo" {
  project_id          = dbtcloud_project.example.id
  schema              = "your_schema"
  user                = "your_user"
  password_wo         = var.teradata_password
  password_wo_version = 1
}
