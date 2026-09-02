package bigquery_credential

import "github.com/hashicorp/terraform-plugin-framework/types"

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

// optionalString maps an empty API value to null, so that fields the API does not return
// don't show up as an empty string in state
func optionalString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}
