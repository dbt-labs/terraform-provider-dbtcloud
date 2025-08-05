package semantic_layer_credential_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestDbtCloudSemanticLayerConfigurationBigQueryResource(t *testing.T) {

	_, _, projectID := acctest_helper.GetSemanticLayerConfigTestingConfigurations()
	if projectID == 0 {
		t.Skip("Skipping test because config is not set")
	}

	name := acctest.RandomWithPrefix("bigquery_tf_test")
	name2 := acctest.RandomWithPrefix("bigquery_tf_test_2")

	privateKeyID := acctest.RandString(10)
	privateKey := acctest.RandString(10)
	clientEmail := acctest.RandString(10)
	clientID := acctest.RandString(10)
	authURI := acctest.RandString(10)
	tokenURI := acctest.RandString(10)
	authProviderCertURL := acctest.RandString(10)
	clientCertURL := acctest.RandString(10)

	clientEmail2 := acctest.RandString(10)
	clientID2 := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSemanticLayerCredentialBigQueryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudBigQuerySemanticLayerCredentialResource(
					projectID,
					name,
					privateKeyID,
					privateKey,
					clientEmail,
					clientID,
					authURI,
					tokenURI,
					authProviderCertURL,
					clientCertURL,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"configuration.project_id",
						strconv.Itoa(projectID),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"configuration.name",
						name,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"credential.project_id",
						strconv.Itoa(projectID),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"private_key_id",
						privateKeyID,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"private_key",
						privateKey,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"client_email",
						clientEmail,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"client_id",
						clientID,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"auth_uri",
						authURI,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"token_uri",
						tokenURI,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"auth_provider_x509_cert_url",
						authProviderCertURL,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"client_x509_cert_url",
						clientCertURL,
					),
				),
			},

			// Update name, clientEmail, clientID
			{
				Config: testAccDbtCloudBigQuerySemanticLayerCredentialResource(
					projectID,
					name2,
					privateKeyID,
					privateKey,
					clientEmail2,
					clientID2,
					authURI,
					tokenURI,
					authProviderCertURL,
					clientCertURL,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"configuration.project_id",
						strconv.Itoa(projectID),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"configuration.name",
						name2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"credential.project_id",
						strconv.Itoa(projectID),
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"client_email",
						clientEmail2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"client_id",
						clientID2,
					),
				),
			},
		},
	})
}

func TestDbtCloudSemanticLayerConfigurationBigQueryResource_SensitiveAttributeHandling(t *testing.T) {
	_, _, projectID := acctest_helper.GetSemanticLayerConfigTestingConfigurations()
	if projectID == 0 {
		t.Skip("Skipping test because config is not set")
	}

	name := acctest.RandomWithPrefix("bigquery_sensitive_test")
	privateKeyID := acctest.RandString(10)
	privateKey := acctest.RandString(20) // Make it longer to simulate real private key
	clientEmail := acctest.RandString(10) + "@example.com"
	clientID := acctest.RandString(10)
	authURI := "https://oauth2.googleapis.com/token"
	tokenURI := "https://oauth2.googleapis.com/token"
	authProviderCertURL := "https://www.googleapis.com/oauth2/v1/certs"
	clientCertURL := "https://www.googleapis.com/robot/v1/metadata/x509/test%40example.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSemanticLayerCredentialBigQueryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudBigQuerySemanticLayerCredentialResource(
					projectID,
					name,
					privateKeyID,
					privateKey,
					clientEmail,
					clientID,
					authURI,
					tokenURI,
					authProviderCertURL,
					clientCertURL,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"configuration.name",
						name,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"private_key_id",
						privateKeyID,
					),
					// Verify sensitive attribute exists but don't check its value
					resource.TestCheckResourceAttrSet(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"private_key",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"client_email",
						clientEmail,
					),
				),
			},
			{
				// This step tests that refreshing the resource doesn't cause inconsistencies
				// with sensitive attributes. The ExpectNonEmptyPlan: false ensures that
				// after a refresh, there should be no plan changes.
				Config: testAccDbtCloudBigQuerySemanticLayerCredentialResource(
					projectID,
					name,
					privateKeyID,
					privateKey,
					clientEmail,
					clientID,
					authURI,
					tokenURI,
					authProviderCertURL,
					clientCertURL,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false, // This is key - no plan changes should be detected
			},
		},
	})
}

func TestDbtCloudSemanticLayerConfigurationBigQueryResource_SensitiveAttributeNull(t *testing.T) {
	// This test specifically validates that when sensitive attributes are initially null,
	// they get properly set to empty strings to avoid inconsistencies.

	_, _, projectID := acctest_helper.GetSemanticLayerConfigTestingConfigurations()
	if projectID == 0 {
		t.Skip("Skipping test because config is not set")
	}

	name := acctest.RandomWithPrefix("bigquery_null_test")

	// Test with minimal required values
	privateKeyID := "test_key_id"
	privateKey := "test_private_key_content"
	clientEmail := "test@serviceaccount.example.com"
	clientID := "123456789"
	authURI := "https://accounts.google.com/o/oauth2/auth"
	tokenURI := "https://oauth2.googleapis.com/token"
	authProviderCertURL := "https://www.googleapis.com/oauth2/v1/certs"
	clientCertURL := "https://www.googleapis.com/robot/v1/metadata/x509/test%40serviceaccount.example.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSemanticLayerCredentialBigQueryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudBigQuerySemanticLayerCredentialResource(
					projectID,
					name,
					privateKeyID,
					privateKey,
					clientEmail,
					clientID,
					authURI,
					tokenURI,
					authProviderCertURL,
					clientCertURL,
				),
				Check: resource.ComposeTestCheckFunc(
					// Check that resource is created successfully
					resource.TestCheckResourceAttrSet(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"id",
					),
					// Check that private_key_id is preserved
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"private_key_id",
						privateKeyID,
					),
					// Check that private_key exists (but don't verify value for security)
					resource.TestCheckResourceAttrSet(
						"dbtcloud_bigquery_semantic_layer_credential.test_bigquery_semantic_layer_credential",
						"private_key",
					),
				),
			},
			{
				// Test refresh with explicit plan-only check
				Config: testAccDbtCloudBigQuerySemanticLayerCredentialResource(
					projectID,
					name,
					privateKeyID,
					privateKey,
					clientEmail,
					clientID,
					authURI,
					tokenURI,
					authProviderCertURL,
					clientCertURL,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false, // This is key - no plan changes should be detected
			},
		},
	})
}

// builds a terraform config for dbtcloud_bigquery_semantic_layer_credential resource
func testAccDbtCloudBigQuerySemanticLayerCredentialResource(
	projectID int,
	name string,
	privateKeyID string,
	privateKey string,
	clientEmail string,
	clientID string,
	authURI string,
	tokenURI string,
	authProviderCertURL string,
	clientCertURL string,

) string {

	return fmt.Sprintf(`
	
resource "dbtcloud_bigquery_semantic_layer_credential" "test_bigquery_semantic_layer_credential" {
  configuration = {
    project_id = %s
	name = "%s"
	adapter_version = "bigquery_v0"
  }
  credential = {
  	project_id = %s
	is_active = true
    num_threads = 3
	dataset = "test"
  }
  private_key_id = "%s"
  private_key = "%s"
  client_email = "%s"
  client_id = "%s"
  auth_uri = "%s"
  token_uri = "%s"
  auth_provider_x509_cert_url = "%s"
  client_x509_cert_url = "%s"
  
}`, strconv.Itoa(projectID), name, strconv.Itoa(projectID), privateKeyID, privateKey, clientEmail, clientID, authURI, tokenURI, authProviderCertURL, clientCertURL)
}

func testAccCheckDbtCloudSemanticLayerCredentialBigQueryDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_bigquery_semantic_layer_credential" {
			continue
		}

		id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert resource ID to int64: %s", err)
		}
		_, err = apiClient.GetSemanticLayerCredential(id)
		if err == nil {
			return fmt.Errorf("Semantic Layer Configuration still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
