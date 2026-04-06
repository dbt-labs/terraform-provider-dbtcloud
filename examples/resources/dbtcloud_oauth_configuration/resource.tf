// Using the classic sensitive attribute (stored in state)
resource "dbtcloud_oauth_configuration" "entra" {
  type               = "entra"
  name               = "My Entra ID Oauth integration"
  client_id          = "client-id"
  client_secret      = "client-secret"
  redirect_uri       = "http://example.com"
  token_url          = "http://example.com"
  authorize_url      = "http://example.com"
  application_id_uri = "uri"
}

resource "dbtcloud_oauth_configuration" "okta" {
  type          = "okta"
  name          = "My Okta Oauth integration"
  client_id     = "client-id"
  client_secret = "client-secret"
  redirect_uri  = "http://example.com"
  token_url     = "http://example.com"
  authorize_url = "http://example.com"
}

// Using write-only attributes (not stored in state, requires Terraform >= 1.11)
//
// The client_secret_wo value is never persisted in the Terraform state file.
// Use client_secret_wo_version to trigger an update when the client secret changes.
variable "oauth_client_secret" {
  type      = string
  ephemeral = true
}

resource "dbtcloud_oauth_configuration" "entra_wo" {
  type                    = "entra"
  name                    = "My Entra ID Oauth integration"
  client_id               = "client-id"
  client_secret_wo        = var.oauth_client_secret
  client_secret_wo_version = 1
  redirect_uri            = "http://example.com"
  token_url               = "http://example.com"
  authorize_url           = "http://example.com"
  application_id_uri      = "uri"
}
