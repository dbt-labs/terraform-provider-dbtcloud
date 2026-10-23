package analytics_credential_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const resourceAddress = "dbtcloud_analytics_credential.test"

// stateHasNoSecret fails when any attribute of the state carries one of the given values
func stateHasNoSecret(secrets ...string) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, res := range s.RootModule().Resources {
			if res.Type != "dbtcloud_analytics_credential" {
				continue
			}
			for key, value := range res.Primary.Attributes {
				for _, secret := range secrets {
					if strings.Contains(value, secret) {
						return fmt.Errorf("the state attribute %s carries a secret", key)
					}
				}
			}
		}
		return nil
	}
}

func configOf(t *testing.T, call map[string]interface{}) map[string]interface{} {
	t.Helper()
	config, ok := call["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("no config in the payload: %v", call)
	}
	return config
}

// Rule: secrets never reach the state.
// given: a snowflake credential created with password_wo
// when: it is applied
// then: the POST body carries the password, the state has no secret value and the schema has no plain password attribute
func TestAnalyticsCredential_PasswordWriteOnlySentButNotStored(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "snowflake_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
		Config: snowflakeConfig(srv.URL, 42, passwordAttributes),
		Check: tfresource.ComposeTestCheckFunc(
			func(*terraform.State) error {
				posts := calls(srv, t, "POST", collectionPath)
				if len(posts) != 1 {
					return fmt.Errorf("expected one POST, got %d", len(posts))
				}
				if got := configOf(t, posts[0])["password"]; got != "super-secret-password" {
					return fmt.Errorf("expected the password in the POST body, got %v", got)
				}
				return nil
			},
			tfresource.TestCheckNoResourceAttr(resourceAddress, "snowflake.password_wo"),
			tfresource.TestCheckNoResourceAttr(resourceAddress, "snowflake.password"),
			tfresource.TestCheckResourceAttr(resourceAddress, "snowflake.password_wo_version", "1"),
			stateHasNoSecret("super-secret-password"),
		),
	}}))
}

// Rule: secrets never reach the state.
// given: a snowflake credential created with private_key_wo and private_key_passphrase_wo
// when: it is applied
// then: the POST body carries both, the state has neither and the schema has no plain secret attribute
func TestAnalyticsCredential_KeypairWriteOnlySentButNotStored(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "snowflake_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
		Config: snowflakeConfig(srv.URL, 42, keypairAttributes),
		Check: tfresource.ComposeTestCheckFunc(
			func(*terraform.State) error {
				posts := calls(srv, t, "POST", collectionPath)
				if len(posts) != 1 {
					return fmt.Errorf("expected one POST, got %d", len(posts))
				}
				config := configOf(t, posts[0])
				if config["private_key"] != "super-secret-key" {
					return fmt.Errorf("expected the private key in the POST body, got %v", config["private_key"])
				}
				if config["private_key_passphrase"] != "super-secret-passphrase" {
					return fmt.Errorf("expected the passphrase in the POST body, got %v", config["private_key_passphrase"])
				}
				if _, ok := config["password"]; ok {
					return fmt.Errorf("keypair auth must not send a password")
				}
				return nil
			},
			tfresource.TestCheckNoResourceAttr(resourceAddress, "snowflake.private_key_wo"),
			tfresource.TestCheckNoResourceAttr(resourceAddress, "snowflake.private_key"),
			tfresource.TestCheckNoResourceAttr(resourceAddress, "snowflake.private_key_passphrase_wo"),
			tfresource.TestCheckNoResourceAttr(resourceAddress, "snowflake.private_key_passphrase"),
			stateHasNoSecret("super-secret-key", "super-secret-passphrase"),
		),
	}}))
}

// Rule: secrets never reach the state.
// given: an applied credential
// when: it is refreshed against a GET that deliberately returns password, private_key and private_key_passphrase
// then: none of them is written to the state
func TestAnalyticsCredential_RefreshIgnoresSecretsInResponse(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "snowflake_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{
		{Config: snowflakeConfig(srv.URL, 42, passwordAttributes)},
		{
			PreConfig:    func() { api.mu.Lock(); api.leakSecrets = true; api.mu.Unlock() },
			RefreshState: true,
			Check: tfresource.ComposeTestCheckFunc(
				tfresource.TestCheckNoResourceAttr(resourceAddress, "snowflake.password"),
				tfresource.TestCheckNoResourceAttr(resourceAddress, "snowflake.private_key"),
				tfresource.TestCheckNoResourceAttr(resourceAddress, "snowflake.private_key_passphrase"),
				stateHasNoSecret("leaked-password", "leaked-private-key", "leaked-passphrase"),
			),
		},
	}))
}

// Rule: a secret is only sent again when its version changes.
// given: an applied credential
// when: warehouse changes alone, then password_wo_version is bumped
// then: the first PATCH body has no secret key and the second carries the password
func TestAnalyticsCredential_UpdateSendsSecretOnlyOnVersionBump(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "snowflake_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	withWarehouse := func(warehouse string, version int) string {
		return snowflakeConfig(srv.URL, 42, fmt.Sprintf(`auth_type = "password"
    user = "analytics_user"
    warehouse = %q
    password_wo = "super-secret-password"
    password_wo_version = %d`, warehouse, version))
	}

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{
		{Config: withWarehouse("wh_one", 1)},
		{
			Config: withWarehouse("wh_two", 1),
			Check: func(*terraform.State) error {
				patches := calls(srv, t, "PATCH", itemPath(credentialID))
				if len(patches) != 1 {
					return fmt.Errorf("expected one PATCH, got %d", len(patches))
				}
				config := configOf(t, patches[0])
				for _, key := range secretKeys {
					if _, ok := config[key]; ok {
						return fmt.Errorf("the PATCH body carries %s", key)
					}
				}
				if config["warehouse"] != "wh_two" {
					return fmt.Errorf("expected the new warehouse, got %v", config["warehouse"])
				}
				return nil
			},
		},
		{
			Config: withWarehouse("wh_two", 2),
			Check: func(*terraform.State) error {
				patches := calls(srv, t, "PATCH", itemPath(credentialID))
				if len(patches) != 2 {
					return fmt.Errorf("expected two PATCH calls, got %d", len(patches))
				}
				if got := configOf(t, patches[1])["password"]; got != "super-secret-password" {
					return fmt.Errorf("expected the password after the version bump, got %v", got)
				}
				return nil
			},
		},
	}))
}

// Rule: secrets never reach the state.
// given: a databricks credential created with token_wo
// when: it is applied and refreshed against a GET that returns a token
// then: the POST body carries the token and the state never does; an update without a version bump omits it from the PATCH
func TestAnalyticsCredential_DatabricksTokenWriteOnly(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "databricks_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{
		{
			Config: databricksConfig(srv.URL, 42, databricksAttributes),
			Check: tfresource.ComposeTestCheckFunc(
				func(*terraform.State) error {
					posts := calls(srv, t, "POST", collectionPath)
					if len(posts) != 1 {
						return fmt.Errorf("expected one POST, got %d", len(posts))
					}
					if got := configOf(t, posts[0])["token"]; got != "super-secret-token" {
						return fmt.Errorf("expected the token in the POST body, got %v", got)
					}
					return nil
				},
				tfresource.TestCheckNoResourceAttr(resourceAddress, "databricks.token_wo"),
				tfresource.TestCheckNoResourceAttr(resourceAddress, "databricks.token"),
				tfresource.TestCheckResourceAttr(resourceAddress, "databricks.catalog", "main"),
				stateHasNoSecret("super-secret-token"),
			),
		},
		{
			PreConfig:    func() { api.mu.Lock(); api.leakSecrets = true; api.mu.Unlock() },
			RefreshState: true,
			Check: tfresource.ComposeTestCheckFunc(
				tfresource.TestCheckNoResourceAttr(resourceAddress, "databricks.token"),
				stateHasNoSecret("leaked-token"),
			),
		},
		{
			PreConfig: func() { api.mu.Lock(); api.leakSecrets = false; api.mu.Unlock() },
			Config: databricksConfig(srv.URL, 42, `token_wo = "super-secret-token"
    token_wo_version = 1
    catalog = "other"`),
			Check: func(*terraform.State) error {
				patches := calls(srv, t, "PATCH", itemPath(credentialID))
				if len(patches) != 1 {
					return fmt.Errorf("expected one PATCH, got %d", len(patches))
				}
				config := configOf(t, patches[0])
				if _, ok := config["token"]; ok {
					return fmt.Errorf("the PATCH body carries the token without a version bump")
				}
				if config["catalog"] != "other" {
					return fmt.Errorf("expected the new catalog, got %v", config["catalog"])
				}
				return nil
			},
		},
	}))
}

// Rule: a secret is only sent again when its version changes.
// given: an applied databricks credential
// when: token_wo_version is bumped
// then: the PATCH body carries the token
func TestAnalyticsCredential_DatabricksTokenSentOnVersionBump(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "databricks_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{
		{Config: databricksConfig(srv.URL, 42, databricksAttributes)},
		{
			Config: databricksConfig(srv.URL, 42, `token_wo = "rotated-token"
    token_wo_version = 2
    catalog = "main"`),
			Check: func(*terraform.State) error {
				patches := calls(srv, t, "PATCH", itemPath(credentialID))
				if len(patches) != 1 {
					return fmt.Errorf("expected one PATCH, got %d", len(patches))
				}
				if got := configOf(t, patches[0])["token"]; got != "rotated-token" {
					return fmt.Errorf("expected the rotated token, got %v", got)
				}
				return nil
			},
		},
	}))
}

// Rule: secrets never reach the state.
// given: a redshift credential created with password_wo
// when: it is applied and refreshed against a GET that returns a password
// then: the POST carries the password, the state does not, and an update without a version bump omits it
func TestAnalyticsCredential_RedshiftPasswordWriteOnly(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "redshift_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{
		{
			Config: redshiftConfig(srv.URL, 42, redshiftAttributes),
			Check: tfresource.ComposeTestCheckFunc(
				func(*terraform.State) error {
					posts := calls(srv, t, "POST", collectionPath)
					if len(posts) != 1 {
						return fmt.Errorf("expected one POST, got %d", len(posts))
					}
					config := configOf(t, posts[0])
					if config["password"] != "super-secret-redshift" || config["username"] != "analytics_user" {
						return fmt.Errorf("unexpected POST config: %v", config)
					}
					return nil
				},
				tfresource.TestCheckNoResourceAttr(resourceAddress, "redshift.password_wo"),
				tfresource.TestCheckNoResourceAttr(resourceAddress, "redshift.password"),
				stateHasNoSecret("super-secret-redshift"),
			),
		},
		{
			PreConfig:    func() { api.mu.Lock(); api.leakSecrets = true; api.mu.Unlock() },
			RefreshState: true,
			Check: tfresource.ComposeTestCheckFunc(
				tfresource.TestCheckNoResourceAttr(resourceAddress, "redshift.password"),
				stateHasNoSecret("leaked-password"),
			),
		},
		{
			PreConfig: func() { api.mu.Lock(); api.leakSecrets = false; api.mu.Unlock() },
			Config: redshiftConfig(srv.URL, 42, `username = "other_user"
    password_wo = "super-secret-redshift"
    password_wo_version = 1`),
			Check: func(*terraform.State) error {
				patches := calls(srv, t, "PATCH", itemPath(credentialID))
				if len(patches) != 1 {
					return fmt.Errorf("expected one PATCH, got %d", len(patches))
				}
				config := configOf(t, patches[0])
				if _, ok := config["password"]; ok {
					return fmt.Errorf("the PATCH body carries the password without a version bump")
				}
				if config["username"] != "other_user" {
					return fmt.Errorf("expected the new username, got %v", config["username"])
				}
				return nil
			},
		},
		{
			Config: redshiftConfig(srv.URL, 42, `username = "other_user"
    password_wo = "rotated-redshift"
    password_wo_version = 2`),
			Check: func(*terraform.State) error {
				patches := calls(srv, t, "PATCH", itemPath(credentialID))
				if len(patches) != 2 {
					return fmt.Errorf("expected two PATCH calls, got %d", len(patches))
				}
				if got := configOf(t, patches[1])["password"]; got != "rotated-redshift" {
					return fmt.Errorf("expected the rotated password, got %v", got)
				}
				return nil
			},
		},
	}))
}

// Rule: secrets never reach the state.
// given: a bigquery service account credential created with private_key_wo
// when: it is applied, refreshed against a GET that returns a private key, and updated
// then: the POST carries the key, the state never does, and an update without a version bump omits it
func TestAnalyticsCredential_BigQueryPrivateKeyWriteOnly(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "bigquery_v1"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	withProject := func(project string) string {
		return bigqueryConfig(srv.URL, 42, serviceAccountFields+"\n    execution_project = \""+project+"\"")
	}

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{
		{
			Config: withProject("billing-one"),
			Check: tfresource.ComposeTestCheckFunc(
				func(*terraform.State) error {
					posts := calls(srv, t, "POST", collectionPath)
					if len(posts) != 1 {
						return fmt.Errorf("expected one POST, got %d", len(posts))
					}
					if got := configOf(t, posts[0])["private_key"]; got != "super-secret-bq-key" {
						return fmt.Errorf("expected the private key in the POST body, got %v", got)
					}
					return nil
				},
				tfresource.TestCheckNoResourceAttr(resourceAddress, "bigquery.private_key_wo"),
				tfresource.TestCheckNoResourceAttr(resourceAddress, "bigquery.private_key"),
				tfresource.TestCheckNoResourceAttr(resourceAddress, "bigquery.auth_type"),
				stateHasNoSecret("super-secret-bq-key"),
			),
		},
		{
			PreConfig:    func() { api.mu.Lock(); api.leakSecrets = true; api.mu.Unlock() },
			RefreshState: true,
			Check: tfresource.ComposeTestCheckFunc(
				tfresource.TestCheckNoResourceAttr(resourceAddress, "bigquery.private_key"),
				stateHasNoSecret("leaked-private-key"),
			),
		},
		{
			PreConfig: func() { api.mu.Lock(); api.leakSecrets = false; api.mu.Unlock() },
			Config:    withProject("billing-two"),
			Check: func(*terraform.State) error {
				patches := calls(srv, t, "PATCH", itemPath(credentialID))
				if len(patches) != 1 {
					return fmt.Errorf("expected one PATCH, got %d", len(patches))
				}
				config := configOf(t, patches[0])
				if _, ok := config["private_key"]; ok {
					return fmt.Errorf("the PATCH body carries the private key without a version bump")
				}
				if config["execution_project"] != "billing-two" {
					return fmt.Errorf("expected the new execution project, got %v", config["execution_project"])
				}
				return nil
			},
		},
	}))
}

// Rule: secrets never reach the state.
// given: a bigquery workload identity federation credential
// when: it is applied
// then: the POST carries the path and no service account key, the state has no secret, and nothing is re-sent when unchanged
func TestAnalyticsCredential_BigQueryWifHasNoSecret(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "bigquery_v1"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	config := bigqueryConfig(srv.URL, 42, `auth_type = "external-oauth-wif"
    `+wifFields)

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{
		{
			Config: config,
			Check: tfresource.ComposeTestCheckFunc(
				func(*terraform.State) error {
					posts := calls(srv, t, "POST", collectionPath)
					if len(posts) != 1 {
						return fmt.Errorf("expected one POST, got %d", len(posts))
					}
					body := configOf(t, posts[0])
					if body["workload_pool_provider_path"] == nil || body["auth_type"] != "external-oauth-wif" {
						return fmt.Errorf("unexpected POST config: %v", body)
					}
					for _, key := range []string{"private_key", "private_key_id", "client_email", "client_id", "auth_uri", "token_uri"} {
						if _, ok := body[key]; ok {
							return fmt.Errorf("the WIF body carries %s", key)
						}
					}
					return nil
				},
				tfresource.TestCheckResourceAttr(resourceAddress, "bigquery.auth_type", "external-oauth-wif"),
				tfresource.TestCheckNoResourceAttr(resourceAddress, "bigquery.private_key_wo"),
			),
		},
		{Config: config, PlanOnly: true},
	}))
}
