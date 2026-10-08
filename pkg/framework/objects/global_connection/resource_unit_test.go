package global_connection_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func bigqueryPlanOnlyConfig(hostURL, privateKeyLines string) string {
	return fmt.Sprintf(`
provider "dbtcloud" {
  host_url   = "%s"
  token      = "dummy-token"
  account_id = 1
}

resource "dbtcloud_global_connection" "test" {
  name = "test-bigquery"

  bigquery = {
    gcp_project_id              = "my-gcp-project-id"
    private_key_id              = "my-private-key-id"
    %s
    client_email                = "my_client_email"
    client_id                   = "my_client_id"
    auth_uri                    = "my_auth_uri"
    token_uri                   = "my_token_uri"
    auth_provider_x509_cert_url = "my_auth_provider_x509_cert_url"
    client_x509_cert_url        = "my_client_x509_cert_url"
  }
}`, hostURL, privateKeyLines)
}

func setupOfflineProvider(t *testing.T) (hostURL string, cleanup func()) {
	t.Helper()
	originalTFAcc := os.Getenv("TF_ACC")
	os.Setenv("TF_ACC", "1")

	// the mock server only provides a valid host_url, PlanOnly steps never reach the API
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{})
	return srv.URL, func() {
		srv.Close()
		if originalTFAcc == "" {
			os.Unsetenv("TF_ACC")
		} else {
			os.Setenv("TF_ACC", originalTFAcc)
		}
	}
}

// The validators on `private_key` point at their write-only sibling with a relative path.
// This guards against "Invalid Path Expression for Schema" errors (see #712), which only
// show up when the validators actually run, so the plaintext key has to be set.
func TestGlobalConnectionBigQuery_PrivateKeyValidatesPlan(t *testing.T) {
	hostURL, cleanup := setupOfflineProvider(t)
	defer cleanup()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bigqueryPlanOnlyConfig(
					hostURL,
					`private_key = "ABCDEFGHIJKL"`,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// write-only attributes require Terraform >= 1.11
func TestGlobalConnectionBigQuery_PrivateKeyWriteOnlyValidatesPlan(t *testing.T) {
	hostURL, cleanup := setupOfflineProvider(t)
	defer cleanup()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		Steps: []resource.TestStep{
			{
				// only the write-only key: accepted by BigQueryAuthValidator
				Config: bigqueryPlanOnlyConfig(
					hostURL,
					`private_key_wo = "ABCDEFGHIJKL"
    private_key_wo_version = 1`,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// both keys: rejected
				Config: bigqueryPlanOnlyConfig(
					hostURL,
					`private_key = "ABCDEFGHIJKL"
    private_key_wo = "ABCDEFGHIJKL"
    private_key_wo_version = 1`,
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)private_key.*private_key_wo|Invalid Attribute Combination`),
			},
			{
				// neither key: rejected (the service account checks only run when the auth type is set)
				Config: bigqueryPlanOnlyConfig(
					hostURL,
					`deployment_env_auth_type = "service-account-json"`,
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be specified for BigQuery`),
			},
		},
	})
}
