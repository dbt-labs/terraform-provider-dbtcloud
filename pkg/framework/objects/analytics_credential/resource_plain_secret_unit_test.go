package analytics_credential_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const plainPasswordAttributes = `auth_type = "password"
    user = "analytics_user"
    warehouse = "analytics_wh"
    password = "plain-password"`

const plainDatabricksAttributes = `token = "plain-token"
    catalog = "main"`

const plainRedshiftAttributes = `username = "analytics_user"
    password = "plain-redshift"`

const plainServiceAccountFields = `private_key_id = "key-id"
    private_key = "plain-bq-key"
    client_email = "sa@example.iam.gserviceaccount.com"
    client_id = "123"
    auth_uri = "https://accounts.google.com/o/oauth2/auth"
    token_uri = "https://oauth2.googleapis.com/token"
    auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
    client_x509_cert_url = "https://www.googleapis.com/robot/v1/metadata/x509/sa"`

// Rule: a plain secret is Sensitive and stored in the state, and it reaches the API like a write-only one.
// given: a credential of each adapter created with a plain secret
// when: it is applied
// then: the POST body carries the secret and the state holds it.
func TestAnalyticsCredential_PlainSecretIsSentAndStored(t *testing.T) {
	cases := []struct {
		name      string
		adapter   string
		config    func(hostURL string) string
		apiKey    string
		stateKey  string
		secret    string
		blockName string
	}{
		{"snowflake password", "snowflake_v0", func(u string) string { return snowflakeConfig(u, 42, plainPasswordAttributes) }, "password", "snowflake.password", "plain-password", "snowflake"},
		{"databricks token", "databricks_v0", func(u string) string { return databricksConfig(u, 42, plainDatabricksAttributes) }, "token", "databricks.token", "plain-token", "databricks"},
		{"redshift password", "redshift_v0", func(u string) string { return redshiftConfig(u, 42, plainRedshiftAttributes) }, "password", "redshift.password", "plain-redshift", "redshift"},
		{"bigquery private key", "bigquery_v1", func(u string) string { return bigqueryConfig(u, 42, plainServiceAccountFields) }, "private_key", "bigquery.private_key", "plain-bq-key", "bigquery"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setupAcc(t)
			api := newFakeAPI(map[int]string{42: tc.adapter})
			srv := testhelpers.SetupMockServer(t, api.handlers())
			defer srv.Close()

			tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
				Config: tc.config(srv.URL),
				Check: tfresource.ComposeTestCheckFunc(
					func(*terraform.State) error {
						posts := calls(srv, t, "POST", collectionPath)
						if len(posts) != 1 {
							return fmt.Errorf("expected one POST, got %d", len(posts))
						}
						if got := configOf(t, posts[0])[tc.apiKey]; got != tc.secret {
							return fmt.Errorf("expected the secret in the POST body, got %v", got)
						}
						return nil
					},
					tfresource.TestCheckResourceAttr(resourceAddress, tc.stateKey, tc.secret),
				),
			}}))
		})
	}
}

// Rule: a secret is given either as a plain value or as a write-only one, never both.
// given: a snowflake block with password and password_wo
// when: the plan is made
// then: the plan fails on the conflict.
func TestAnalyticsCredential_PlainAndWriteOnlySecretConflict(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "snowflake_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
		Config:      snowflakeConfig(srv.URL, 42, plainPasswordAttributes+"\n    password_wo = \"other\"\n    password_wo_version = 1"),
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
	}}))
}

// Rule: a secret is required, as a plain value or a write-only one.
// given: a Databricks block and a Redshift block with neither form of the secret
// when: the plan is made
// then: the plan fails saying the secret is missing.
func TestAnalyticsCredential_SecretIsRequiredInEitherForm(t *testing.T) {
	cases := map[string]struct {
		adapter string
		config  func(hostURL string) string
	}{
		"databricks": {"databricks_v0", func(u string) string { return databricksConfig(u, 42, `catalog = "main"`) }},
		"redshift":   {"redshift_v0", func(u string) string { return redshiftConfig(u, 42, `username = "analytics_user"`) }},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			setupAcc(t)
			api := newFakeAPI(map[int]string{42: tc.adapter})
			srv := testhelpers.SetupMockServer(t, api.handlers())
			defer srv.Close()

			tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
				Config:      tc.config(srv.URL),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)missing required|invalid attribute combination|at least one`),
			}}))
		})
	}
}

// Rule: a plain secret is sent when it changes and left out otherwise.
// given: a snowflake credential created with a plain password
// when: an unrelated attribute is changed, then the password is changed
// then: the first update carries no password and the second carries the new one.
func TestAnalyticsCredential_PlainSecretIsSentOnlyWhenChanged(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "snowflake_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	patchedPassword := func(index int) (interface{}, error) {
		patches := calls(srv, t, "PATCH", itemPath(credentialID))
		if len(patches) <= index {
			return nil, fmt.Errorf("expected at least %d PATCH calls, got %d", index+1, len(patches))
		}
		return configOf(t, patches[index])["password"], nil
	}

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{
		{Config: snowflakeConfig(srv.URL, 42, plainPasswordAttributes)},
		{
			Config: snowflakeConfig(srv.URL, 42, `auth_type = "password"
    user = "analytics_user"
    warehouse = "other_wh"
    password = "plain-password"`),
			Check: func(*terraform.State) error {
				got, err := patchedPassword(0)
				if err != nil {
					return err
				}
				if got != nil {
					return fmt.Errorf("an unchanged password was sent: %v", got)
				}
				return nil
			},
		},
		{
			Config: snowflakeConfig(srv.URL, 42, `auth_type = "password"
    user = "analytics_user"
    warehouse = "other_wh"
    password = "rotated-password"`),
			Check: func(*terraform.State) error {
				got, err := patchedPassword(1)
				if err != nil {
					return err
				}
				if got != "rotated-password" {
					return fmt.Errorf("expected the rotated password in the PATCH body, got %v", got)
				}
				return nil
			},
		},
	}))
}

// Rule: the secrets allowed in a block follow its auth type, in both forms.
// given: Snowflake and BigQuery blocks with a plain secret that fits, and with one that does not fit the auth type
// when: the plan is made
// then: the fitting secret plans and the other fails as not allowed.
func TestAnalyticsCredential_PlainSecretsFollowTheAuthType(t *testing.T) {
	notAllowed := regexp.MustCompile(`not allowed`)
	cases := []struct {
		name        string
		adapter     string
		config      func(hostURL string) string
		expectError *regexp.Regexp
	}{
		{"keypair with a plain private key", "snowflake_v0", func(u string) string {
			return snowflakeConfig(u, 42, "auth_type = \"keypair\"\n    user = \"u\"\n    warehouse = \"w\"\n    private_key = \"plain-key\"")
		}, nil},
		{"keypair with a plain password", "snowflake_v0", func(u string) string {
			return snowflakeConfig(u, 42, "auth_type = \"keypair\"\n    user = \"u\"\n    warehouse = \"w\"\n    private_key = \"plain-key\"\n    password = \"plain-password\"")
		}, notAllowed},
		{"password with a plain private key", "snowflake_v0", func(u string) string {
			return snowflakeConfig(u, 42, plainPasswordAttributes+"\n    private_key = \"plain-key\"")
		}, notAllowed},
		{"workload identity federation with a plain private key", "bigquery_v1", func(u string) string {
			return bigqueryConfig(u, 42, "auth_type = \"external-oauth-wif\"\n    "+wifFields+"\n    private_key = \"plain-key\"")
		}, notAllowed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setupAcc(t)
			api := newFakeAPI(map[int]string{42: tc.adapter})
			srv := testhelpers.SetupMockServer(t, api.handlers())
			defer srv.Close()

			step := tfresource.TestStep{Config: tc.config(srv.URL), PlanOnly: true}
			if tc.expectError != nil {
				step.ExpectError = tc.expectError
			} else {
				step.ExpectNonEmptyPlan = true
			}
			tfresource.Test(t, unitTestCase([]tfresource.TestStep{step}))
		})
	}
}

// Rule: a secret that was set cannot be cleared, in either form.
// given: a keypair credential with a plain passphrase
// when: the passphrase is removed from the configuration
// then: the plan fails saying the API cannot clear it.
func TestAnalyticsCredential_RemovingPlainPassphraseFailsPlan(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "snowflake_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	withPassphrase := "auth_type = \"keypair\"\n    user = \"u\"\n    warehouse = \"w\"\n    private_key = \"plain-key\"\n    private_key_passphrase = \"plain-passphrase\""
	withoutPassphrase := "auth_type = \"keypair\"\n    user = \"u\"\n    warehouse = \"w\"\n    private_key = \"plain-key\""
	tfresource.Test(t, unitTestCase([]tfresource.TestStep{
		{Config: snowflakeConfig(srv.URL, 42, withPassphrase)},
		{
			Config:      snowflakeConfig(srv.URL, 42, withoutPassphrase),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile(`cannot clear`),
		},
	}))
}
