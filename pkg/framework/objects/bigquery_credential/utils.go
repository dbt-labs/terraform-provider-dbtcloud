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

type wifField struct {
	name  string
	value types.String
}

// wifFields returns the v1-only WIF attributes, in schema order
func wifFields(model BigqueryCredentialResourceModel) []wifField {
	return []wifField{
		{"auth_type", model.AuthType},
		{"workload_pool_provider_path", model.WorkloadPoolProviderPath},
		{"service_account_impersonation_url", model.ServiceAccountImpersonationURL},
	}
}
