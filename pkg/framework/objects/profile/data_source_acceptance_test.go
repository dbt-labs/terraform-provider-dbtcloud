package profile_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudProfileDataSource(t *testing.T) {
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	profileKey := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudProfileDataSourceConfig(projectName, profileKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_profile.test",
						"profile_id",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_profile.test",
						"key",
						profileKey,
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_profile.test",
						"connection_id",
					),
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_profile.test",
						"credentials_id",
					),
				),
			},
		},
	})
}

func testAccDbtCloudProfileDataSourceConfig(projectName, profileKey string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_global_connection" "test_connection" {
  name = "profile_ds_test_connection_%s"

  snowflake = {
    account   = "test-account"
    warehouse = "test-warehouse"
    database  = "test-database"
  }
}

resource "dbtcloud_snowflake_credential" "test_credential" {
  is_active  = true
  project_id = dbtcloud_project.test_project.id
  auth_type  = "password"
  database   = "test-database"
  role       = "test-role"
  warehouse  = "test-warehouse"
  schema     = "test_schema"
  user       = "test-user"
  password   = "test-password"
  num_threads = 3
}

resource "dbtcloud_profile" "test_profile" {
  project_id     = dbtcloud_project.test_project.id
  key            = "%s"
  connection_id  = dbtcloud_global_connection.test_connection.id
  credentials_id = dbtcloud_snowflake_credential.test_credential.credential_id
}

data "dbtcloud_profile" "test" {
  project_id = dbtcloud_profile.test_profile.project_id
  profile_id = dbtcloud_profile.test_profile.profile_id
}
`, projectName, projectName, profileKey)
}
