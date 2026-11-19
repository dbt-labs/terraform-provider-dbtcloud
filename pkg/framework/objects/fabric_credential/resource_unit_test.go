package fabric_credential

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveAuthentication(t *testing.T) {
	// Rule: A configured authentication wins, and an unset one is inferred from the user
	//
	// given: a configured, an unknown and a null authentication, with and without a user
	// when: the effective authentication is resolved
	// then: the configured value is kept and the others follow the user
	cases := []struct {
		name       string
		configured types.String
		user       string
		want       string
	}{
		{"configured external_oauth", types.StringValue("external_oauth"), "", "external_oauth"},
		{"configured value beats the user", types.StringValue("ServicePrincipal"), "someone", "ServicePrincipal"},
		{"unknown with a user", types.StringUnknown(), "someone", "ActiveDirectoryPassword"},
		{"null with a user", types.StringNull(), "someone", "ActiveDirectoryPassword"},
		{"null without a user", types.StringNull(), "", "ServicePrincipal"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveAuthentication(tc.configured, tc.user); got != tc.want {
				t.Errorf("resolveAuthentication = %q, want %q", got, tc.want)
			}
		})
	}
}
