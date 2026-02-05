package salesforce_credential_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudSalesforceCredentialDataSource(t *testing.T) {
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	username := "test_user@example.com"
	clientID := "test_client_id"
	privateKey := "-----BEGIN RSA PRIVATE KEY-----MIIEpAIBAAKCAQEA" + strings.ToUpper(acctest.RandStringFromCharSet(50, acctest.CharSetAlpha)) + "-----END RSA PRIVATE KEY-----"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudSalesforceCredentialDataSourceConfig(
					projectName,
					connectionName,
					username,
					clientID,
					privateKey,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_salesforce_credential.test",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_salesforce_credential.test",
						"credential_id",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_salesforce_credential.test",
						"project_id",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_salesforce_credential.test",
						"username",
						username,
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_salesforce_credential.test",
						"target_name",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_salesforce_credential.test",
						"num_threads",
					),
				),
			},
		},
	})
}

func testAccDbtCloudSalesforceCredentialDataSourceConfig(
	projectName string,
	connectionName string,
	username string,
	clientID string,
	privateKey string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test" {
  name = "%s"
}

resource "dbtcloud_global_connection" "test" {
  name = "%s"
  salesforce = {
    login_url                  = "https://login.salesforce.com"
    database                   = "default"
    data_transform_run_timeout = 300
  }
}

resource "dbtcloud_salesforce_credential" "test" {
  project_id  = dbtcloud_project.test.id
  username    = "%s"
  client_id   = "%s"
  private_key = "%s"
}

resource "dbtcloud_environment" "prod" {
  name            = "Production"
  type            = "deployment"
  dbt_version     = "latest"
  project_id      = dbtcloud_project.test.id
  deployment_type = "production"
  credential_id   = dbtcloud_salesforce_credential.test.credential_id
  connection_id   = dbtcloud_global_connection.test.id
}

data "dbtcloud_salesforce_credential" "test" {
  project_id    = dbtcloud_project.test.id
  credential_id = dbtcloud_salesforce_credential.test.credential_id
}
`, projectName, connectionName, username, clientID, privateKey)
}
