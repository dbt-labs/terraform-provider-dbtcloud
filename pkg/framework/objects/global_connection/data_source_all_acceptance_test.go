package global_connection_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudGlobalConnectionsDatasource(t *testing.T) {
	connectionName := acctest.RandStringFromCharSet(19, acctest.CharSetAlphaNum)

	// Unparallelized because of flakiness
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudGlobalConnectionsDatasourceBasicConfig(
					connectionName,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_global_connections.test",
						"connections.0.id",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_global_connections.test",
						"connections.0.name",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_global_connections.test",
						"connections.0.adapter_version",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_global_connections.test",
						"connections.0.environment__count",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_global_connections.test",
						"connections.1.id",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_global_connections.test",
						"connections.1.name",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_global_connections.test",
						"connections.1.adapter_version",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_global_connections.test",
						"connections.1.environment__count",
					),
				),
			},
		},
	})
}

func testAccDbtCloudGlobalConnectionsDatasourceBasicConfig(
	connectionName string,
) string {

	return fmt.Sprintf(`

resource dbtcloud_global_connection connection1 {
  name = "%[1]s1"

  snowflake = {
    account = "account"
    warehouse = "warehouse"
    database = "database"
    allow_sso = true
    client_session_keep_alive = false
	role = "role"
  }
}

resource dbtcloud_global_connection connection2 {
  name = "%[1]s2"

  bigquery = {
    gcp_project_id              = "my-gcp-project-id"
    timeout_seconds             = 1000
    private_key_id              = "my-private-key-id"
    private_key                 = "ABCDEFGHIJKL"
    client_email                = "my_client_email"
    client_id                   = "my_client_id"
    auth_uri                    = "my_auth_uri"
    token_uri                   = "my_token_uri"
    auth_provider_x509_cert_url = "my_auth_provider_x509_cert_url"
    client_x509_cert_url        = "my_client_x509_cert_url"
    application_id              = "oauth_application_id"
    application_secret          = "oauth_secret_id"

  }
}

data dbtcloud_global_connections test {
  depends_on = [dbtcloud_global_connection.connection1, dbtcloud_global_connection.connection2]
}

`, connectionName)
}
