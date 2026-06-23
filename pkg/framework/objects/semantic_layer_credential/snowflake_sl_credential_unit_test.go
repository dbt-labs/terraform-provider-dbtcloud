package semantic_layer_credential_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestSnowflakeSLCredential_NestedSchemaValidatesPlan is an offline regression
// test for the "Invalid Path Expression for Schema" bug.
//
// The Snowflake (and Redshift/Postgres/Databricks) Semantic Layer credential
// resources reuse the stand-alone credential schema verbatim under a nested
// `credential` attribute. The reused schema declared its write-only conflict
// validators with absolute paths (path.MatchRoot("private_key_wo"), etc). When
// nested, those paths resolve against the Semantic Layer resource root — where
// `private_key_wo` does not exist (it lives at `credential.private_key_wo`) — so
// ValidateResourceConfig rejected every plan with an "Invalid Path Expression for
// Schema" error before any API call was made.
//
// These steps only run validate + plan (PlanOnly), which is exactly where the
// bug fired, so no create/read API mocks are required.
//
// Note on what gets set: the broken validators are attached to the *plaintext*
// attributes (password / private_key / private_key_passphrase) and merely point
// at their write-only sibling (password_wo, ...) as the conflict target. The
// ConflictsWith validator short-circuits when the attribute it is attached to is
// null, so it only resolves the (previously invalid) sibling path when the
// plaintext attribute has a value. We therefore set the plaintext secrets — not
// the *_wo attributes — to force those validators to execute against the nested
// schema and reproduce the original "Invalid Path Expression for Schema" error.
func TestSnowflakeSLCredential_NestedSchemaValidatesPlan(t *testing.T) {
	originalTFAcc := os.Getenv("TF_ACC")
	os.Setenv("TF_ACC", "1")
	defer func() {
		if originalTFAcc == "" {
			os.Unsetenv("TF_ACC")
		} else {
			os.Setenv("TF_ACC", originalTFAcc)
		}
	}()

	// A mock server is only needed so the provider has a valid host_url; PlanOnly
	// steps never reach create/read, so no handlers are registered.
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{})
	defer srv.Close()

	providerConfig := fmt.Sprintf(`
		provider "dbtcloud" {
			host_url   = "%s"
			token      = "dummy-token"
			account_id = 1
		}`, srv.URL)

	passwordAuthConfig := providerConfig + `
		resource "dbtcloud_snowflake_semantic_layer_credential" "test" {
		  configuration = {
		    project_id      = 123
		    name            = "test-sl-credential"
		    adapter_version = "snowflake_v0"
		  }
		  credential = {
		    project_id                = 123
		    is_active                 = true
		    auth_type                 = "password"
		    role                      = "test_role"
		    warehouse                 = "test_warehouse"
		    user                      = "test_user"
		    password                  = "test_password"
		    num_threads               = 3
		    semantic_layer_credential = true
		  }
		}`

	keypairAuthConfig := providerConfig + `
		resource "dbtcloud_snowflake_semantic_layer_credential" "test" {
		  configuration = {
		    project_id      = 123
		    name            = "test-sl-credential"
		    adapter_version = "snowflake_v0"
		  }
		  credential = {
		    project_id                = 123
		    is_active                 = true
		    auth_type                 = "keypair"
		    role                      = "test_role"
		    warehouse                 = "test_warehouse"
		    private_key               = "test_private_key"
		    private_key_passphrase    = "test_passphrase"
		    num_threads               = 3
		    semantic_layer_credential = true
		  }
		}`

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Exercises the password / password_wo conflict validators.
				Config:             passwordAuthConfig,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Exercises the private_key(_passphrase) / *_wo conflict validators.
				Config:             keypairAuthConfig,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
