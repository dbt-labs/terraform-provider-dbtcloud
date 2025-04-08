package teradata_credential_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudStarburstCredentialDataSource(t *testing.T) {
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schema := "test_schema"
	user := "test_user"
	password := "test_password"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudTeradataCredentialDataSourceConfig(
					projectName,
					schema,
					user,
					password,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_teradata_credential.test",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_teradata_credential.test",
						"credential_id",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_teradata_credential.test",
						"schema",
						schema,
					),
				),
			},
		},
	})
}

func testAccDbtCloudTeradataCredentialDataSourceConfig(
	projectName string,
	schema string,
	user string,
	password string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test" {
  name = "%s"
}

resource "dbtcloud_teradata_credential" "test" {
  project_id           = dbtcloud_project.test.id
  schema               = "%s"
  user                 = "%s"
  password             = "%s"
}

data "dbtcloud_teradata_credential" "test" {
  project_id    = dbtcloud_project.test.id
  credential_id = dbtcloud_teradata_credential.test.credential_id
}
`, projectName, schema, user, password)
}
