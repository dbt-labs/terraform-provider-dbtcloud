// Using the classic sensitive attributes (stored in state)
resource "dbtcloud_athena_credential" "example" {
  project_id            = dbtcloud_project.example.id
  aws_access_key_id     = "your-access-key-id"
  aws_secret_access_key = "your-secret-access-key"
  schema                = "your_schema"
}

// Using write-only attributes (not stored in state, requires Terraform >= 1.11)
//
// The aws_access_key_id_wo and aws_secret_access_key_wo values are never persisted in the Terraform state file.
// Use aws_access_key_id_wo_version / aws_secret_access_key_wo_version to trigger an update when the secret changes.
variable "athena_aws_access_key_id" {
  type      = string
  ephemeral = true
}

variable "athena_aws_secret_access_key" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_athena_credential" "example_wo" {
  project_id                       = dbtcloud_project.example.id
  aws_access_key_id_wo             = var.athena_aws_access_key_id
  aws_access_key_id_wo_version     = 1
  aws_secret_access_key_wo         = var.athena_aws_secret_access_key
  aws_secret_access_key_wo_version = 1
  schema                           = "your_schema"
}
