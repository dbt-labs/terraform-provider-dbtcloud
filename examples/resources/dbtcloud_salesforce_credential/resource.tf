# Using the classic sensitive attributes (stored in state)
resource "dbtcloud_salesforce_credential" "my_salesforce_cred" {
  project_id  = dbtcloud_project.dbt_project.id
  username    = "user@example.com"
  client_id   = "your-oauth-client-id"
  private_key = "private-key value"
  target_name = "default"
  num_threads = 6
}

// Using write-only attributes (not stored in state, requires Terraform >= 1.11)
//
// The client_id_wo and private_key_wo values are never persisted in the Terraform state file.
// Use client_id_wo_version / private_key_wo_version to trigger an update when the secret changes.
variable "salesforce_client_id" {
  type      = string
  ephemeral = true
}

variable "salesforce_private_key" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_salesforce_credential" "my_salesforce_cred_wo" {
  project_id             = dbtcloud_project.dbt_project.id
  username               = "user@example.com"
  client_id_wo           = var.salesforce_client_id
  client_id_wo_version   = 1
  private_key_wo         = var.salesforce_private_key
  private_key_wo_version = 1
  target_name            = "default"
  num_threads            = 6
}
