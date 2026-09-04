package openai_integration_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccDbtCloudOpenAIIntegrationResource_OpenAI tests the full lifecycle of
// an openai key type: create, update (key rotation), import, destroy.
func TestAccDbtCloudOpenAIIntegrationResource_OpenAI(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest_helper.TestAccPreCheck(t)
			sweepOpenAIIntegrations(t)
		},
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudOpenAIIntegrationDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccDbtCloudOpenAIIntegrationOpenAIConfig("sk-test-key-v1", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_openai_integration.test",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_openai_integration.test",
						"account_id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"key_type",
						"openai",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_openai_integration.test",
						"created_at",
					),
				),
			},
			// Rotate the key — increment key_value_wo_version
			{
				Config: testAccDbtCloudOpenAIIntegrationOpenAIConfig("sk-test-key-v2", 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"key_type",
						"openai",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"key_value_wo_version",
						"2",
					),
				),
			},
			// Import — key_value_wo is write-only and never stored in state;
			// key_value_wo_version will be absent after import.
			{
				ResourceName:            "dbtcloud_openai_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key_value_wo", "key_value_wo_version"},
			},
		},
	})
}

// TestAccDbtCloudOpenAIIntegrationResource_KeyValue tests using the regular
// sensitive key_value attribute (for older Terraform versions).
func TestAccDbtCloudOpenAIIntegrationResource_KeyValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest_helper.TestAccPreCheck(t)
			sweepOpenAIIntegrations(t)
		},
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudOpenAIIntegrationDestroy,
		Steps: []resource.TestStep{
			// Create with sensitive key_value
			{
				Config: testAccDbtCloudOpenAIIntegrationKeyValueConfig("sk-test-sensitive-v1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"key_type",
						"openai",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"key_value",
						"sk-test-sensitive-v1",
					),
				),
			},
			// Update key_value in place
			{
				Config: testAccDbtCloudOpenAIIntegrationKeyValueConfig("sk-test-sensitive-v2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"key_value",
						"sk-test-sensitive-v2",
					),
				),
			},
			// Import — key_value is never returned by the API; must be ignored.
			{
				ResourceName:            "dbtcloud_openai_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key_value"},
			},
		},
	})
}

// TestAccDbtCloudOpenAIIntegrationResource_AzureOpenAI tests the azure_openai
// key type including all required Azure-specific fields.
func TestAccDbtCloudOpenAIIntegrationResource_AzureOpenAI(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest_helper.TestAccPreCheck(t)
			sweepOpenAIIntegrations(t)
		},
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudOpenAIIntegrationDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccDbtCloudOpenAIIntegrationAzureConfig(
					"az-test-key-v1", 1,
					"https://my-deployment.openai.azure.com/",
					"gpt-4o",
					"2024-02-01",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"key_type",
						"azure_openai",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"azure_endpoint",
						"https://my-deployment.openai.azure.com/",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"azure_deployment_name",
						"gpt-4o",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"azure_api_version",
						"2024-02-01",
					),
				),
			},
			// Update deployment name and rotate key
			{
				Config: testAccDbtCloudOpenAIIntegrationAzureConfig(
					"az-test-key-v2", 2,
					"https://my-deployment.openai.azure.com/",
					"gpt-4o-mini",
					"2024-02-01",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"azure_deployment_name",
						"gpt-4o-mini",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"key_value_wo_version",
						"2",
					),
				),
			},
			// Import
			{
				ResourceName:            "dbtcloud_openai_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key_value_wo", "key_value_wo_version"},
			},
		},
	})
}

// TestAccDbtCloudOpenAIIntegrationResource_KeyTypeSwitch tests switching
// key_type from openai to azure_openai in place.
func TestAccDbtCloudOpenAIIntegrationResource_KeyTypeSwitch(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest_helper.TestAccPreCheck(t)
			sweepOpenAIIntegrations(t)
		},
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudOpenAIIntegrationDestroy,
		Steps: []resource.TestStep{
			// Start as openai
			{
				Config: testAccDbtCloudOpenAIIntegrationOpenAIConfig("sk-test-key-v1", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"key_type",
						"openai",
					),
				),
			},
			// Switch to azure_openai — same resource, in-place PATCH
			{
				Config: testAccDbtCloudOpenAIIntegrationAzureConfig(
					"az-test-key-v1", 1,
					"https://my-deployment.openai.azure.com/",
					"gpt-4o",
					"2024-02-01",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_openai_integration.test",
						"key_type",
						"azure_openai",
					),
				),
			},
		},
	})
}

// TestAccDbtCloudOpenAIIntegrationResource_ValidationErrors tests that
// ValidateConfig correctly rejects invalid configurations.
func TestAccDbtCloudOpenAIIntegrationResource_ValidationErrors(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		// every step fails in config validation, so this one never reaches the API
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// openai without any key
			{
				Config:      testAccDbtCloudOpenAIIntegrationNoKeyConfig("openai"),
				ExpectError: regexp.MustCompile(`Missing API key`),
			},
			// azure_openai without azure fields
			{
				Config:      testAccDbtCloudOpenAIIntegrationNoKeyConfig("azure_openai"),
				ExpectError: regexp.MustCompile(`Missing API key|Missing required field`),
			},
			// azure_openai without key but with azure fields
			{
				Config:      testAccDbtCloudOpenAIIntegrationAzureNoKeyConfig(),
				ExpectError: regexp.MustCompile(`Missing API key`),
			},
			// openai with azure fields set
			{
				Config:      testAccDbtCloudOpenAIIntegrationOpenAIWithAzureFieldsConfig(),
				ExpectError: regexp.MustCompile(`must not be set when key_type is openai`),
			},
			// both key_value and key_value_wo set
			{
				Config:      testAccDbtCloudOpenAIIntegrationBothKeysConfig(),
				ExpectError: regexp.MustCompile(`Attribute "key_value_wo" cannot be specified when "key_value" is specified`),
			},
		},
	})
}

// ── destroy check ─────────────────────────────────────────────────────────────

func testAccCheckDbtCloudOpenAIIntegrationDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("issue getting the client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_openai_integration" {
			continue
		}

		id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("could not parse OpenAI integration ID %q: %w", rs.Primary.ID, err)
		}

		_, err = apiClient.GetOpenAIIntegration(id)
		if err == nil {
			return fmt.Errorf("OpenAI integration %d still exists after destroy", id)
		}

		notFoundErr := regexp.MustCompile("resource-not-found")
		if !notFoundErr.MatchString(err.Error()) {
			return fmt.Errorf("unexpected error checking OpenAI integration destroy: %w", err)
		}
	}

	return nil
}

// ── config helpers ────────────────────────────────────────────────────────────

func testAccDbtCloudOpenAIIntegrationOpenAIConfig(key string, version int) string {
	return fmt.Sprintf(`
resource "dbtcloud_openai_integration" "test" {
  key_type             = "openai"
  key_value_wo         = %q
  key_value_wo_version = %d
}
`, key, version)
}

func testAccDbtCloudOpenAIIntegrationKeyValueConfig(key string) string {
	return fmt.Sprintf(`
resource "dbtcloud_openai_integration" "test" {
  key_type  = "openai"
  key_value = %q
}
`, key)
}

func testAccDbtCloudOpenAIIntegrationAzureConfig(
	key string,
	version int,
	endpoint, deploymentName, apiVersion string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_openai_integration" "test" {
  key_type              = "azure_openai"
  key_value_wo          = %q
  key_value_wo_version  = %d
  azure_endpoint        = %q
  azure_deployment_name = %q
  azure_api_version     = %q
}
`, key, version, endpoint, deploymentName, apiVersion)
}

func testAccDbtCloudOpenAIIntegrationNoKeyConfig(keyType string) string {
	return fmt.Sprintf(`
resource "dbtcloud_openai_integration" "test" {
  key_type = %q
}
`, keyType)
}

func testAccDbtCloudOpenAIIntegrationAzureNoKeyConfig() string {
	return `
resource "dbtcloud_openai_integration" "test" {
  key_type              = "azure_openai"
  azure_endpoint        = "https://my-deployment.openai.azure.com/"
  azure_deployment_name = "gpt-4o"
  azure_api_version     = "2024-02-01"
}
`
}

func testAccDbtCloudOpenAIIntegrationOpenAIWithAzureFieldsConfig() string {
	return `
resource "dbtcloud_openai_integration" "test" {
  key_type              = "openai"
  key_value_wo          = "sk-test"
  key_value_wo_version  = 1
  azure_endpoint        = "https://my-deployment.openai.azure.com/"
  azure_deployment_name = "gpt-4o"
  azure_api_version     = "2024-02-01"
}
`
}

func testAccDbtCloudOpenAIIntegrationBothKeysConfig() string {
	return `
resource "dbtcloud_openai_integration" "test" {
  key_type             = "openai"
  key_value            = "sk-sensitive"
  key_value_wo         = "sk-writeonly"
  key_value_wo_version = 1
}
`
}

// sweepOpenAIIntegrations deletes integrations left behind by an earlier run. The API
// allows one per account, so a leftover makes every create in this file fail on the
// account_id unique constraint - which the API reports as a 500 - until it is removed.
// A run that crashes or whose destroy step fails leaves exactly that behind, so the
// failure repeats on every later run.
func sweepOpenAIIntegrations(t *testing.T) {
	t.Helper()

	client, err := acctest_helper.SharedClient()
	if err != nil {
		t.Fatalf("Issue getting the client: %s", err)
	}

	integrations, err := client.GetAllOpenAIIntegrations()
	if err != nil {
		// Best effort: the test below reports the real failure if a leftover is blocking it.
		t.Logf("could not list the existing OpenAI integrations: %s", err)
		return
	}

	for _, integration := range integrations {
		if integration.ID == nil {
			continue
		}
		if err := client.DeleteOpenAIIntegration(*integration.ID); err != nil {
			t.Logf("could not delete the leftover OpenAI integration %d: %s", *integration.ID, err)
			continue
		}
		t.Logf("deleted OpenAI integration %d left behind by an earlier run", *integration.ID)
	}
}
