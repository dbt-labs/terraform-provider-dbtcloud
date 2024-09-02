package global_connection_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudGlobalConnectionDatasource(t *testing.T) {
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	oAuthClientID := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	oAuthClientSecret := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudGlobalConnectionDatasourceBasicConfig(
					connectionName,
					oAuthClientID,
					oAuthClientSecret,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_global_connection.test",
						"adapter_version",
						"snowflake_v0",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_global_connection.test",
						"name",
						connectionName,
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_global_connection.test",
						"snowflake.account",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_global_connection.test",
						"snowflake.database",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_global_connection.test",
						"snowflake.warehouse",
					),
				),
			},
		},
	})

}

func testAccDbtCloudGlobalConnectionDatasourceBasicConfig(
	connectionName, oAuthClientID, oAuthClientSecret string,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  snowflake = {
    account = "account"
    warehouse = "warehouse"
    database = "database"
    allow_sso = true
    oauth_client_id = "%s"
    oauth_client_secret = "%s"
    client_session_keep_alive = false
	role = "role"
  }
}

data dbtcloud_global_connection test {
  id = dbtcloud_global_connection.test.id
}

`, connectionName, oAuthClientID, oAuthClientSecret)
}
