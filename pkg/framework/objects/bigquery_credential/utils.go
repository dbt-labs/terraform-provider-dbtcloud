package bigquery_credential

const (
	AuthTypeServiceAccountJSON = "service-account-json"
	AuthTypeOAuthSecrets       = "oauth-secrets"
	AuthTypeExternalOAuthWIF   = "external-oauth-wif"
)

var AuthTypes = []string{
	AuthTypeServiceAccountJSON,
	AuthTypeOAuthSecrets,
	AuthTypeExternalOAuthWIF,
}
