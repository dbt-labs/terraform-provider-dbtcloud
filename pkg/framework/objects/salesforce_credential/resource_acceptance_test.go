package salesforce_credential_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudSalesforceCredentialResource(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	username := strings.ToLower(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)) + "@example.com"
	username2 := strings.ToLower(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)) + "@example.com"
	clientID := strings.ToUpper(acctest.RandStringFromCharSet(20, acctest.CharSetAlpha))
	privateKey := "-----BEGIN RSA PRIVATE KEY-----MIIEpAIBAAKCAQEA" + strings.ToUpper(acctest.RandStringFromCharSet(50, acctest.CharSetAlpha)) + "-----END RSA PRIVATE KEY-----"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSalesforceCredentialDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDbtCloudSalesforceCredentialResourceConfig(
					projectName,
					connectionName,
					username,
					clientID,
					privateKey,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudSalesforceCredentialExists(
						"dbtcloud_salesforce_credential.test_credential",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_salesforce_credential.test_credential",
						"username",
						username,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_salesforce_credential.test_credential",
						"target_name",
						"default",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_salesforce_credential.test_credential",
						"num_threads",
						"6",
					),
				),
			},
			// Update testing
			{
				Config: testAccDbtCloudSalesforceCredentialResourceConfig(
					projectName,
					connectionName,
					username2,
					clientID,
					privateKey,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudSalesforceCredentialExists(
						"dbtcloud_salesforce_credential.test_credential",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_salesforce_credential.test_credential",
						"username",
						username2,
					),
				),
			},
			// Import testing
			{
				ResourceName:            "dbtcloud_salesforce_credential.test_credential",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_id", "private_key"},
			},
		},
	})
}

func testAccDbtCloudSalesforceCredentialResourceConfig(
	projectName, connectionName, username, clientID, privateKey string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_global_connection" "salesforce" {
  name = "%s"
  salesforce = {
    login_url                  = "https://login.salesforce.com"
    database                   = "default"
    data_transform_run_timeout = 300
  }
}

resource "dbtcloud_salesforce_credential" "test_credential" {
  project_id  = dbtcloud_project.test_project.id
  username    = "%s"
  client_id   = "%s"
  private_key = "%s"
}
`, projectName, connectionName, username, clientID, privateKey)
}

func testAccCheckDbtCloudSalesforceCredentialExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Salesforce credential ID is set")
		}

		return nil
	}
}

func testAccCheckDbtCloudSalesforceCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_salesforce_credential" {
			continue
		}

		idParts := strings.Split(rs.Primary.ID, ":")
		if len(idParts) != 2 {
			return fmt.Errorf("Unexpected ID format: %s", rs.Primary.ID)
		}

		var projectID, credentialID int
		fmt.Sscanf(idParts[0], "%d", &projectID)
		fmt.Sscanf(idParts[1], "%d", &credentialID)

		_, err := apiClient.GetSalesforceCredential(projectID, credentialID)
		if err == nil {
			return fmt.Errorf("Salesforce credential still exists")
		}

		if !strings.HasPrefix(err.Error(), "resource-not-found") {
			return fmt.Errorf("Unexpected error: %s", err.Error())
		}
	}

	return nil
}
