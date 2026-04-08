# Create a SCIM token to allow an identity provider (e.g. Okta, Azure AD) to
# provision and deprovision users and groups in dbt Cloud.
#
# The token value is returned only on creation and stored in Terraform state
# as a sensitive value. It is never returned by subsequent API reads, so
# changing the `name` forces a new token to be created.
resource "dbtcloud_scim_config_token" "okta" {
  name = "okta-scim"
}

# The token string can be passed to other resources or outputs.
# Mark outputs as sensitive when exposing the token value.
output "scim_token" {
  value     = dbtcloud_scim_config_token.okta.token_string
  sensitive = true
}
