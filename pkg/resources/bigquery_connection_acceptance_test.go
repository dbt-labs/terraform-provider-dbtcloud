package resources_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudBigQueryConnectionResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	privateKey := strings.ToUpper(acctest.RandStringFromCharSet(100, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudBigQueryConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudBigQueryConnectionResourceBasicConfig(connectionName, projectName, privateKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists("dbtcloud_bigquery_connection.test_connection"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "name", connectionName),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "type", "bigquery"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "gcp_project_id", "test_project_id"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "timeout_seconds", "1000"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "private_key_id", "test_private_key_id"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "private_key", privateKey),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "client_email", "test_client_email"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "client_id", "test_client_id"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "auth_uri", "test_auth_uri"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "token_uri", "test_token_uri"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "auth_provider_x509_cert_url", "test_auth_provider_x509_cert_url"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "client_x509_cert_url", "test_client_x509_cert_url"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "retries", "3"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "is_configured_for_oauth", "false"),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudBigQueryConnectionResourceBasicConfig(connectionName2, projectName, privateKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists("dbtcloud_bigquery_connection.test_connection"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "name", connectionName2),
				),
			},
			// MODIFY AND ADD OAUTH
			{
				Config: testAccDbtCloudBigQueryConnectionResourceOAuthConfig(connectionName2, projectName, privateKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudConnectionExists("dbtcloud_bigquery_connection.test_connection"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "application_secret", "test_application_secret"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "application_id", "test_application_id"),
					resource.TestCheckResourceAttr("dbtcloud_bigquery_connection.test_connection", "is_configured_for_oauth", "true"),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_bigquery_connection.test_connection",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "application_secret", "application_id"},
			},
		},
	})
}

func testAccDbtCloudBigQueryConnectionResourceBasicConfig(connectionName, projectName, privateKey string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_bigquery_connection" "test_connection" {
  name        = "%s"
  type = "bigquery"
  project_id = dbtcloud_project.test_project.id
  gcp_project_id = "test_project_id"
  timeout_seconds = 1000
  private_key_id = "test_private_key_id"
  private_key = "%s"
  client_email = "test_client_email"
  client_id = "test_client_id"
  auth_uri = "test_auth_uri"
  token_uri = "test_token_uri"
  auth_provider_x509_cert_url = "test_auth_provider_x509_cert_url"
  client_x509_cert_url = "test_client_x509_cert_url"
  retries = 3
}
`, projectName, connectionName, privateKey)
}

func testAccDbtCloudBigQueryConnectionResourceOAuthConfig(connectionName, projectName, privateKey string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}

resource "dbtcloud_bigquery_connection" "test_connection" {
  name        = "%s"
  type = "bigquery"
  project_id = dbtcloud_project.test_project.id
  gcp_project_id = "test_project_id"
  timeout_seconds = 1000
  private_key_id = "test_private_key_id"
  private_key = "%s"
  client_email = "test_client_email"
  client_id = "test_client_id"
  auth_uri = "test_auth_uri"
  token_uri = "test_token_uri"
  auth_provider_x509_cert_url = "test_auth_provider_x509_cert_url"
  client_x509_cert_url = "test_client_x509_cert_url"
  retries = 3
  application_secret = "test_application_secret"
  application_id = "test_application_id"
}
`, projectName, connectionName, privateKey)
}

func testAccCheckDbtCloudBigQueryConnectionDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_bigquery_connection" {
			continue
		}
		projectId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0]
		connectionId := strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1]

		_, err := apiClient.GetConnection(connectionId, projectId)
		if err == nil {
			return fmt.Errorf("Connection still exists")
		}
		notFoundErr := "not found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
