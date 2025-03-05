package athena_credential_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudAthenaCredentialDataSource(t *testing.T) {
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	schema := "test_schema"
	awsAccessKeyID := "test_access_key_id"
	awsSecretAccessKey := "test_secret_access_key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudAthenaCredentialDataSourceConfig(
					projectName,
					schema,
					awsAccessKeyID,
					awsSecretAccessKey,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_athena_credential.test",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_athena_credential.test",
						"credential_id",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_athena_credential.test",
						"schema",
						schema,
					),
				),
			},
		},
	})
}

func testAccDbtCloudAthenaCredentialDataSourceConfig(
	projectName string,
	schema string,
	awsAccessKeyID string,
	awsSecretAccessKey string,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test" {
  name = "%s"
}

resource "dbtcloud_athena_credential" "test" {
  project_id           = dbtcloud_project.test.id
  schema               = "%s"
  aws_access_key_id    = "%s"
  aws_secret_access_key = "%s"
}

data "dbtcloud_athena_credential" "test" {
  project_id    = dbtcloud_project.test.id
  credential_id = dbtcloud_athena_credential.test.credential_id
}
`, projectName, schema, awsAccessKeyID, awsSecretAccessKey)
}
