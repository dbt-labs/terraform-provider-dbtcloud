package azure_ad_application_test

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccDbtCloudAzureADApplicationResource tests create, update, and import
// of an Azure AD application.
//
// Required env vars:
//
//	DBT_ACCEPTANCE_TEST_AZURE_AD_ORG       — Azure DevOps org name
//	DBT_ACCEPTANCE_TEST_AZURE_AD_CLIENT_ID — App registration client ID
//	DBT_ACCEPTANCE_TEST_AZURE_AD_SECRET    — App registration client secret
//	DBT_ACCEPTANCE_TEST_AZURE_AD_TENANT_ID — Azure AD tenant ID
func TestAccDbtCloudAzureADApplicationResource(t *testing.T) {
	orgName := os.Getenv("DBT_ACCEPTANCE_TEST_AZURE_AD_ORG")
	clientID := os.Getenv("DBT_ACCEPTANCE_TEST_AZURE_AD_CLIENT_ID")
	clientSecret := os.Getenv("DBT_ACCEPTANCE_TEST_AZURE_AD_SECRET")
	tenantID := os.Getenv("DBT_ACCEPTANCE_TEST_AZURE_AD_TENANT_ID")

	if orgName == "" || clientID == "" || clientSecret == "" || tenantID == "" {
		t.Skip(
			"Skipping Azure AD application acceptance tests: set DBT_ACCEPTANCE_TEST_AZURE_AD_ORG, " +
				"DBT_ACCEPTANCE_TEST_AZURE_AD_CLIENT_ID, DBT_ACCEPTANCE_TEST_AZURE_AD_SECRET, and " +
				"DBT_ACCEPTANCE_TEST_AZURE_AD_TENANT_ID to run this test.",
		)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudAzureADApplicationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with default auth method
			{
				Config: testAccDbtCloudAzureADApplicationConfig(
					orgName, clientID, clientSecret, tenantID, "service_user",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_azure_ad_application.test",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_azure_ad_application.test",
						"account_id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_azure_ad_application.test",
						"organization_name",
						orgName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_azure_ad_application.test",
						"azure_service_authentication_method",
						"service_user",
					),
				),
			},
			// Step 2: Update auth method to service_principal
			{
				Config: testAccDbtCloudAzureADApplicationConfig(
					orgName, clientID, clientSecret, tenantID, "service_principal",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_azure_ad_application.test",
						"azure_service_authentication_method",
						"service_principal",
					),
				),
			},
			// Step 3: Import
			// client_id, client_secret, tenant_id are not returned by the API
			// and will be empty after import.
			{
				ResourceName:      "dbtcloud_azure_ad_application.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"client_id",
					"client_secret",
					"tenant_id",
				},
			},
		},
	})
}

func testAccDbtCloudAzureADApplicationConfig(
	orgName, clientID, clientSecret, tenantID, authMethod string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_azure_ad_application" "test" {
  organization_name                    = %q
  client_id                            = %q
  client_secret                        = %q
  tenant_id                            = %q
  azure_service_authentication_method  = %q
}
`, orgName, clientID, clientSecret, tenantID, authMethod)
}

func testAccCheckDbtCloudAzureADApplicationDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("issue getting the client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_azure_ad_application" {
			continue
		}

		id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("could not parse Azure AD application ID %q: %w", rs.Primary.ID, err)
		}

		_, err = apiClient.GetAzureADApplication(id)
		if err == nil {
			return fmt.Errorf("Azure AD application %d still exists", id)
		}

		notFoundErr := regexp.MustCompile("resource-not-found")
		if !notFoundErr.MatchString(err.Error()) {
			return fmt.Errorf("expected resource-not-found error, got: %s", err)
		}
	}

	return nil
}
