package platform_metadata_credentials_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// These tests use placeholder keys: dbt Cloud doesn't connect to BigQuery when a platform
// metadata credential is created or updated. catalog_ingestion_enabled stays false in every step,
// because enabling it would start an ingestion run against the placeholder key.

const (
	bqAccResource = "dbtcloud_bigquery_platform_metadata_credential.test"
	// the connection's impersonate_service_account, and the URL dbt Cloud derives from it
	bqAccImpersonatedAccount = "impersonated@my-project.iam.gserviceaccount.com"
	bqAccWorkloadPoolPath    = "//iam.googleapis.com/projects/1/locations/global/workloadIdentityPools/pool/providers/provider"
)

var bqAccImpersonationURLPattern = regexp.MustCompile(
	`^https://iamcredentials\.googleapis\.com/v1/projects/-/serviceAccounts/` +
		regexp.QuoteMeta(bqAccImpersonatedAccount) + `:generateAccessToken$`,
)

// bqAccConnection returns a BigQuery global connection. v1 selects the latest adapter
// (bigquery_v1), which is the only one that supports WIF; otherwise the legacy bigquery_v0 is used.
func bqAccConnection(name string, v1 bool, impersonate string) string {
	extra := fmt.Sprintf("\n    use_latest_adapter = %t", v1)
	if impersonate != "" {
		extra += fmt.Sprintf("\n    impersonate_service_account = %q", impersonate)
	}
	return fmt.Sprintf(`
resource "dbtcloud_global_connection" "bq" {
  name = "%s"

  bigquery = {
    gcp_project_id              = "my-gcp-project-id"
    private_key_id              = "placeholder"
    private_key                 = "placeholder"
    client_email                = "placeholder@example.com"
    client_id                   = "placeholder"
    auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
    token_uri                   = "https://oauth2.googleapis.com/token"
    auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
    client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/placeholder"%s
  }
}
`, name, extra)
}

func bqAccCredential(body string) string {
	return fmt.Sprintf(`
resource "dbtcloud_bigquery_platform_metadata_credential" "test" {
  connection_id = dbtcloud_global_connection.bq.id
%s
}
`, body)
}

func bqAccServiceAccount(keyLines, keyID, extra string) string {
	return fmt.Sprintf(`
  private_key_id              = %q
  client_email                = "sa@my-project.iam.gserviceaccount.com"
  client_id                   = "1234567890"
  auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
  token_uri                   = "https://oauth2.googleapis.com/token"
  auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
  client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/sa"
  %s
  %s
`, keyID, keyLines, extra)
}

func bqAccWIF(extra string) string {
	return fmt.Sprintf(`
  auth_type                   = "external-oauth-wif"
  workload_pool_provider_path = %q
  %s
`, bqAccWorkloadPoolPath, extra)
}

func testAccCheckDbtCloudBigQueryPlatformMetadataCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("issue getting the client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_bigquery_platform_metadata_credential" {
			continue
		}

		credentialID, err := strconv.ParseInt(rs.Primary.Attributes["credential_id"], 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert credential_id to int64: %w", err)
		}

		_, err = apiClient.GetPlatformMetadataCredential(credentialID)
		if err == nil {
			return fmt.Errorf("BigQuery platform metadata credential still exists")
		}
		if !strings.Contains(err.Error(), "resource-not-found") {
			return fmt.Errorf("expected resource-not-found, got %s", err)
		}
	}

	return nil
}

// credentialIDSame checks that a step updated the credential in place.
func credentialIDSame(id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		current := s.RootModule().Resources[bqAccResource].Primary.Attributes["credential_id"]
		if *id == "" {
			*id = current
			return nil
		}
		if current != *id {
			return fmt.Errorf("credential_id changed from %s to %s, it was replaced", *id, current)
		}
		return nil
	}
}

func bqAccName() string {
	return strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
}

// Service account key on bigquery_v0 with auth_type left out: create, toggle a flag, rotate the
// key, import. v0 never stores auth_type, so this also guards against a diff on every plan.
func TestAccDbtCloudBigQueryPlatformMetadataCredentialResource(t *testing.T) {
	connection := bqAccConnection(bqAccName(), false, "")
	credentialID := ""

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudBigQueryPlatformMetadataCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: connection + bqAccCredential(bqAccServiceAccount(`private_key = "first-key"`, "key-id-1", "")),
				Check: resource.ComposeTestCheckFunc(
					credentialIDSame(&credentialID),
					resource.TestCheckResourceAttrSet(bqAccResource, "id"),
					resource.TestCheckResourceAttr(bqAccResource, "auth_type", "service-account-json"),
					resource.TestCheckResourceAttr(bqAccResource, "adapter_version", "bigquery_v0"),
					resource.TestCheckResourceAttr(bqAccResource, "private_key_id", "key-id-1"),
					resource.TestCheckResourceAttr(bqAccResource, "catalog_ingestion_enabled", "false"),
					resource.TestCheckNoResourceAttr(bqAccResource, "workload_pool_provider_path"),
					resource.TestCheckNoResourceAttr(bqAccResource, "service_account_impersonation_url"),
				),
			},
			{
				Config: connection + bqAccCredential(bqAccServiceAccount(
					`private_key = "first-key"`, "key-id-1", "cost_insights_enabled = true",
				)),
				Check: resource.ComposeTestCheckFunc(
					credentialIDSame(&credentialID),
					resource.TestCheckResourceAttr(bqAccResource, "cost_insights_enabled", "true"),
				),
			},
			{
				// rotate the key
				Config: connection + bqAccCredential(bqAccServiceAccount(
					`private_key = "second-key"`, "key-id-2", "cost_insights_enabled = true",
				)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(bqAccResource, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					credentialIDSame(&credentialID),
					resource.TestCheckResourceAttr(bqAccResource, "private_key_id", "key-id-2"),
				),
			},
			{
				// the private key is never read back
				ResourceName:            bqAccResource,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key"},
			},
		},
	})
}

// WIF on bigquery_v1. dbt Cloud fills in the impersonation URL from the connection's
// impersonate_service_account, and leaves it empty without one. Both must plan empty afterwards.
func TestAccDbtCloudBigQueryPlatformMetadataCredentialResourceWIF(t *testing.T) {
	for name, tc := range map[string]struct {
		impersonate string
		check       resource.TestCheckFunc
	}{
		"with impersonation": {
			bqAccImpersonatedAccount,
			resource.TestMatchResourceAttr(bqAccResource, "service_account_impersonation_url", bqAccImpersonationURLPattern),
		},
		"without impersonation": {
			"",
			resource.TestCheckNoResourceAttr(bqAccResource, "service_account_impersonation_url"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			connection := bqAccConnection(bqAccName(), true, tc.impersonate)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
				CheckDestroy:             testAccCheckDbtCloudBigQueryPlatformMetadataCredentialDestroy,
				Steps: []resource.TestStep{
					{
						Config: connection + bqAccCredential(bqAccWIF("")),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(bqAccResource, "auth_type", "external-oauth-wif"),
							resource.TestCheckResourceAttr(bqAccResource, "adapter_version", "bigquery_v1"),
							resource.TestCheckResourceAttr(bqAccResource, "workload_pool_provider_path", bqAccWorkloadPoolPath),
							resource.TestCheckNoResourceAttr(bqAccResource, "private_key_id"),
							tc.check,
						),
					},
					{
						Config: connection + bqAccCredential(bqAccWIF("cost_insights_enabled = true")),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction(bqAccResource, plancheck.ResourceActionUpdate),
							},
						},
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(bqAccResource, "cost_insights_enabled", "true"),
							tc.check,
						),
					},
					{
						ResourceName:      bqAccResource,
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

// Switching auth_type updates the credential in place. The connection has
// impersonate_service_account, so the URL is filled in on the WIF step and cleared on the way back.
func TestAccDbtCloudBigQueryPlatformMetadataCredentialResourceSwitchAuth(t *testing.T) {
	connection := bqAccConnection(bqAccName(), true, bqAccImpersonatedAccount)
	serviceAccount := connection + bqAccCredential(bqAccServiceAccount(`private_key = "first-key"`, "key-id-1", ""))
	wif := connection + bqAccCredential(bqAccWIF(""))
	credentialID := ""

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudBigQueryPlatformMetadataCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: serviceAccount,
				Check: resource.ComposeTestCheckFunc(
					credentialIDSame(&credentialID),
					resource.TestCheckResourceAttr(bqAccResource, "auth_type", "service-account-json"),
					resource.TestCheckResourceAttr(bqAccResource, "adapter_version", "bigquery_v1"),
					resource.TestCheckNoResourceAttr(bqAccResource, "service_account_impersonation_url"),
				),
			},
			{
				Config: wif,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(bqAccResource, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					credentialIDSame(&credentialID),
					resource.TestCheckResourceAttr(bqAccResource, "auth_type", "external-oauth-wif"),
					resource.TestMatchResourceAttr(bqAccResource, "service_account_impersonation_url", bqAccImpersonationURLPattern),
					resource.TestCheckNoResourceAttr(bqAccResource, "private_key_id"),
				),
			},
			{
				Config: serviceAccount,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(bqAccResource, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					credentialIDSame(&credentialID),
					resource.TestCheckResourceAttr(bqAccResource, "auth_type", "service-account-json"),
					resource.TestCheckResourceAttr(bqAccResource, "private_key_id", "key-id-1"),
					resource.TestCheckNoResourceAttr(bqAccResource, "service_account_impersonation_url"),
				),
			},
		},
	})
}

// write-only attributes require Terraform >= 1.11. The key is rotated the same way on both adapters.
func TestAccDbtCloudBigQueryPlatformMetadataCredentialResourceWriteOnly(t *testing.T) {
	for name, adapter := range map[string]struct {
		v1      bool
		version string
	}{
		"bigquery_v0": {false, "bigquery_v0"},
		"bigquery_v1": {true, "bigquery_v1"},
	} {
		t.Run(name, func(t *testing.T) {
			connection := bqAccConnection(bqAccName(), adapter.v1, "")
			config := func(key string, version int) string {
				return connection + bqAccCredential(bqAccServiceAccount(
					fmt.Sprintf("private_key_wo = %q\n  private_key_wo_version = %d", key, version), "key-id-1", "",
				))
			}
			credentialID := ""

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
				ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
				CheckDestroy:             testAccCheckDbtCloudBigQueryPlatformMetadataCredentialDestroy,
				TerraformVersionChecks: []tfversion.TerraformVersionCheck{
					tfversion.SkipBelow(tfversion.Version1_11_0),
				},
				Steps: []resource.TestStep{
					{
						Config: config("first-key", 1),
						Check: resource.ComposeTestCheckFunc(
							credentialIDSame(&credentialID),
							resource.TestCheckResourceAttr(bqAccResource, "adapter_version", adapter.version),
							resource.TestCheckNoResourceAttr(bqAccResource, "private_key_wo"),
							resource.TestCheckNoResourceAttr(bqAccResource, "private_key"),
							resource.TestCheckResourceAttr(bqAccResource, "private_key_wo_version", "1"),
						),
					},
					{
						// bumping the version rotates the key in place
						Config: config("second-key", 2),
						Check: resource.ComposeTestCheckFunc(
							credentialIDSame(&credentialID),
							resource.TestCheckResourceAttr(bqAccResource, "private_key_wo_version", "2"),
						),
					},
					{
						ResourceName:            bqAccResource,
						ImportState:             true,
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{"private_key_wo_version"},
					},
				},
			})
		})
	}
}

// WIF needs bigquery_v1. On bigquery_v0 dbt Cloud validates the config as a service account key
// and rejects it, and the provider adds a hint to the error.
func TestAccDbtCloudBigQueryPlatformMetadataCredentialResourceWIFOnV0(t *testing.T) {
	connection := bqAccConnection(bqAccName(), false, "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudBigQueryPlatformMetadataCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config:      connection + bqAccCredential(bqAccWIF("")),
				ExpectError: regexp.MustCompile(`external-oauth-wif requires a bigquery_v1\s+connection`),
			},
		},
	})
}
