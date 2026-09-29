package clickhouse_credential_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudClickhouseCredentialDataSource(t *testing.T) {
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	user := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudClickhouseCredentialDataSourceConfig(
					projectName,
					connectionName,
					user,
					password,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_clickhouse_credential.test",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_clickhouse_credential.test",
						"credential_id",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_clickhouse_credential.test",
						"project_id",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_clickhouse_credential.test",
						"user",
						user,
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_clickhouse_credential.test",
						"schema",
						"my_schema",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_clickhouse_credential.test",
						"target_name",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_clickhouse_credential.test",
						"threads",
					),
				),
			},
		},
	})
}

func testAccDbtCloudClickhouseCredentialDataSourceConfig(
	projectName string,
	connectionName string,
	user string,
	password string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test" {
  name = "%s"
}

resource "dbtcloud_global_connection" "test" {
  name = "%s"
  clickhouse = {
    host     = "example.com"
    port     = 8443
    database = "default"
  }
}

resource "dbtcloud_clickhouse_credential" "test" {
  project_id = dbtcloud_project.test.id
  schema     = "my_schema"
  user       = "%s"
  password   = "%s"
}

resource "dbtcloud_environment" "prod" {
  name            = "Production"
  type            = "deployment"
  dbt_version     = "latest"
  project_id      = dbtcloud_project.test.id
  deployment_type = "production"
  credential_id   = dbtcloud_clickhouse_credential.test.credential_id
  connection_id   = dbtcloud_global_connection.test.id
}

data "dbtcloud_clickhouse_credential" "test" {
  project_id    = dbtcloud_project.test.id
  credential_id = dbtcloud_clickhouse_credential.test.credential_id
}
`, projectName, connectionName, user, password)
}
