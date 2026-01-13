package bigquery_credential_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var projectName = strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
var dataset = strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

var createCredentialTestStep = resource.TestStep{
	Config: testAccDbtCloudBigQueryCredentialResourceBasicConfig(projectName, dataset),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudBigQueryCredentialExists(
			"dbtcloud_bigquery_credential.test_credential",
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_bigquery_credential.test_credential",
			"dataset",
			dataset,
		),
	),
}

func TestAccDbtCloudBigQueryCredentialResource(t *testing.T) {
	var importStateTestStep = resource.TestStep{
		ResourceName:            "dbtcloud_bigquery_credential.test_credential",
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"password"},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudBigQueryCredentialDestroy,
		Steps: []resource.TestStep{
			createCredentialTestStep,
			// RENAME
			// MODIFY
			importStateTestStep,
		},
	})
}

func testAccDbtCloudBigQueryCredentialResourceBasicConfig(projectName, dataset string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_bigquery_credential" "test_credential" {
    is_active = true
    project_id = dbtcloud_project.test_project.id
    dataset = "%s"
    num_threads = 3
}
`, projectName, dataset)
}

func testAccCheckDbtCloudBigQueryCredentialExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_bigquery_credential",
		)
		if err != nil {
			return err
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		_, err = apiClient.GetBigQueryCredential(projectId, credentialId)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudBigQueryCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_bigquery_credential" {
			continue
		}
		projectId, credentialId, err := helper.SplitIDToInts(
			rs.Primary.ID,
			"dbtcloud_bigquery_credential",
		)
		if err != nil {
			return err
		}

		_, err = apiClient.GetBigQueryCredential(projectId, credentialId)
		if err == nil {
			return fmt.Errorf("BigQuery credential still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}

// TestAccDbtCloudBigQueryCredentialResourceWithConnectionID tests that
// creating a BigQuery credential with connection_id pointing to a global connection
// with use_latest_adapter=true works correctly and auto-detects the adapter version.
func TestAccDbtCloudBigQueryCredentialResourceWithConnectionID(t *testing.T) {
	projectNameV1 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	datasetV1 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionNameV1 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudBigQueryCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudBigQueryCredentialResourceWithConnectionIDConfig(
					projectNameV1,
					datasetV1,
					connectionNameV1,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudBigQueryCredentialExists(
						"dbtcloud_bigquery_credential.test_credential_v1",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_credential.test_credential_v1",
						"dataset",
						datasetV1,
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_bigquery_credential.test_credential_v1",
						"connection_id",
					),
				),
			},
		},
	})
}

func testAccDbtCloudBigQueryCredentialResourceWithConnectionIDConfig(
	projectName, dataset, connectionName string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_global_connection" "test_connection" {
  name = "%s"

  bigquery = {
    gcp_project_id              = "test-gcp-project"
    private_key_id              = "my-private-key-id"
    private_key                 = "ABCDEFGHIJKL"
    client_email                = "test@test-gcp-project.iam.gserviceaccount.com"
    client_id                   = "123456789"
    auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
    token_uri                   = "https://oauth2.googleapis.com/token"
    auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
    client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/test"
    use_latest_adapter          = true
  }
}

resource "dbtcloud_bigquery_credential" "test_credential_v1" {
  is_active     = true
  project_id    = dbtcloud_project.test_project.id
  dataset       = "%s"
  num_threads   = 4
  connection_id = dbtcloud_global_connection.test_connection.id
}
`, projectName, connectionName, dataset)
}

// TestAccDbtCloudBigQueryCredentialResourceWithoutConnectionID tests that
// creating a BigQuery credential without connection_id (legacy behavior) still works.
func TestAccDbtCloudBigQueryCredentialResourceWithoutConnectionID(t *testing.T) {
	projectNameLegacy := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	datasetLegacy := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudBigQueryCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudBigQueryCredentialResourceWithoutConnectionIDConfig(
					projectNameLegacy,
					datasetLegacy,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudBigQueryCredentialExists(
						"dbtcloud_bigquery_credential.test_credential_legacy",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_bigquery_credential.test_credential_legacy",
						"dataset",
						datasetLegacy,
					),
				),
			},
		},
	})
}

func testAccDbtCloudBigQueryCredentialResourceWithoutConnectionIDConfig(
	projectName, dataset string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_bigquery_credential" "test_credential_legacy" {
  is_active   = true
  project_id  = dbtcloud_project.test_project.id
  dataset     = "%s"
  num_threads = 3
}
`, projectName, dataset)
}
