package synapse_credential_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudSynapseCredentialDataSource(t *testing.T) {
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schema := "test_schema"
	user := "test_user"
	password := "test_password"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudSynapseCredentialDataSourceConfig(
					projectName,
					schema,
					user,
					password,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_synapse_credential.test",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_synapse_credential.test",
						"credential_id",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_synapse_credential.test",
						"schema",
						schema,
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_synapse_credential.test",
						"tenant_id",
						"",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_synapse_credential.test",
						"client_id",
						"",
					),
				),
			},
		},
	})
}

func testAccDbtCloudSynapseCredentialDataSourceConfig(
	projectName string,
	schema string,
	user string,
	password string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test" {
  name = "%s"
}

resource "dbtcloud_synapse_credential" "test" {
  project_id      = dbtcloud_project.test.id
  authentication  = "sql"
  user            = "%s"
  password        = "%s"
  schema          = "%s"
  adapter_type    = "synapse"
  schema_authorization = "sp"
}

data "dbtcloud_synapse_credential" "test" {
  project_id    = dbtcloud_project.test.id
  credential_id = dbtcloud_synapse_credential.test.credential_id
}
`, projectName, user, password, schema)
}
