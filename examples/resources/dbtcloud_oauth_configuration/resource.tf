
resource "dbtcloud_oauth_configuration" "test" {
  type               = "entra"
  name               = "My Entra ID Oauth integration"
  client_id          = "client-id"
  client_secret      = "client-secret"
  redirect_uri       = "http://example.com"
  token_url          = "http://example.com"
  authorize_url      = "http://example.com"
  application_id_uri = "uri"
}

resource "dbtcloud_oauth_configuration" "test" {
  type          = "okta"
  name          = "My Okta Oauth integration"
  client_id     = "client-id"
  client_secret = "client-secret"
  redirect_uri  = "http://example.com"
  token_url     = "http://example.com"
  authorize_url = "http://example.com"
}