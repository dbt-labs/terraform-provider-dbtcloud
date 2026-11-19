package dbt_cloud

import "testing"

func fabricAuthenticationOf(t *testing.T, details AdapterCredentialDetails) interface{} {
	t.Helper()
	field, ok := details.Fields["authentication"]
	if !ok {
		t.Fatalf("authentication field missing from the credential details")
	}
	return field.Value
}

func TestGenerateFabricCredentialDetails_InfersAuthenticationWhenUnset(t *testing.T) {
	// Rule: An unset authentication is inferred from the user
	//
	// given: no authentication, and a case with a user and a case without
	// when: the credential details are generated
	// then: the user case is ActiveDirectoryPassword and the other is ServicePrincipal
	cases := []struct {
		name string
		user string
		want string
	}{
		{"with a user", "someone@example.com", "ActiveDirectoryPassword"},
		{"without a user", "", "ServicePrincipal"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			details, err := GenerateFabricCredentialDetails("", tc.user, "pw", "tenant", "client", "secret", "dbt", "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := fabricAuthenticationOf(t, details); got != tc.want {
				t.Errorf("authentication = %v, want %s", got, tc.want)
			}
		})
	}
}

func TestGenerateFabricCredentialDetails_SendsExternalOAuthAsIs(t *testing.T) {
	// Rule: External OAuth is sent as the authentication with no user or service principal values
	//
	// given: the authentication external_oauth and empty user, password and service principal values
	// when: the credential details are generated
	// then: the authentication is external_oauth and those values stay empty
	details, err := GenerateFabricCredentialDetails("external_oauth", "", "", "", "", "", "dbt", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := fabricAuthenticationOf(t, details); got != "external_oauth" {
		t.Errorf("authentication = %v, want external_oauth", got)
	}
	for _, name := range []string{"user", "password", "tenant_id", "client_id", "client_secret"} {
		if got := details.Fields[name].Value; got != "" {
			t.Errorf("%s = %v, want empty", name, got)
		}
	}
}

func TestGenerateFabricCredentialDetails_ListsExternalOAuthAsAnOption(t *testing.T) {
	// Rule: The built-in credential schema offers External OAuth as an authentication option
	//
	// given: the generated credential details
	// when: the authentication options are read
	// then: external_oauth is one of them
	details, err := GenerateFabricCredentialDetails("", "u", "p", "", "", "", "dbt", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, option := range details.Fields["authentication"].Metadata.Options {
		if option.Value == "external_oauth" {
			return
		}
	}
	t.Errorf("external_oauth is not an authentication option")
}
