package connection_catalog_config_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudConnectionCatalogConfigResource(t *testing.T) {
	// Get testing configuration from environment variables
	config := acctest_helper.GetPlatformMetadataCredentialTestingConfigurations()
	if config == nil {
		t.Skip("Skipping test because required environment variables are not set. " +
			"Set ACC_TEST_SNOWFLAKE_ACCOUNT, ACC_TEST_SNOWFLAKE_DATABASE, ACC_TEST_SNOWFLAKE_WAREHOUSE, " +
			"ACC_TEST_SNOWFLAKE_USER, ACC_TEST_SNOWFLAKE_PASSWORD, and ACC_TEST_SNOWFLAKE_ROLE to run this test.")
	}

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with database filters only
			{
				Config: testAccDbtCloudConnectionCatalogConfigResourceConfig(
					connectionName,
					config.SnowflakeAccount,
					config.SnowflakeDatabase,
					config.SnowflakeWarehouse,
					[]string{"analytics", "reporting"},
					nil,
					nil,
					nil,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("dbtcloud_connection_catalog_config.test", "id"),
					resource.TestCheckResourceAttrSet("dbtcloud_connection_catalog_config.test", "connection_id"),
					resource.TestCheckResourceAttr("dbtcloud_connection_catalog_config.test", "database_allow.#", "2"),
					resource.TestCheckResourceAttr("dbtcloud_connection_catalog_config.test", "database_allow.0", "analytics"),
					resource.TestCheckResourceAttr("dbtcloud_connection_catalog_config.test", "database_allow.1", "reporting"),
				),
			},
			// Update - add more filters
			{
				Config: testAccDbtCloudConnectionCatalogConfigResourceConfig(
					connectionName,
					config.SnowflakeAccount,
					config.SnowflakeDatabase,
					config.SnowflakeWarehouse,
					[]string{"analytics"},
					[]string{"staging"},
					[]string{"public"},
					[]string{"temp_*"},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dbtcloud_connection_catalog_config.test", "database_allow.#", "1"),
					resource.TestCheckResourceAttr("dbtcloud_connection_catalog_config.test", "database_deny.#", "1"),
					resource.TestCheckResourceAttr("dbtcloud_connection_catalog_config.test", "schema_allow.#", "1"),
					resource.TestCheckResourceAttr("dbtcloud_connection_catalog_config.test", "table_deny.#", "1"),
				),
			},
			// Import test
			{
				ResourceName:      "dbtcloud_connection_catalog_config.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDbtCloudConnectionCatalogConfigResourceConfig(
	connectionName string,
	snowflakeAccount string,
	snowflakeDatabase string,
	snowflakeWarehouse string,
	databaseAllow []string,
	databaseDeny []string,
	schemaAllow []string,
	tableDeny []string,
) string {
	var databaseAllowStr, databaseDenyStr, schemaAllowStr, tableDenyStr string

	if databaseAllow != nil {
		databaseAllowStr = fmt.Sprintf(`database_allow = ["%s"]`, strings.Join(databaseAllow, `", "`))
	}
	if databaseDeny != nil {
		databaseDenyStr = fmt.Sprintf(`database_deny = ["%s"]`, strings.Join(databaseDeny, `", "`))
	}
	if schemaAllow != nil {
		schemaAllowStr = fmt.Sprintf(`schema_allow = ["%s"]`, strings.Join(schemaAllow, `", "`))
	}
	if tableDeny != nil {
		tableDenyStr = fmt.Sprintf(`table_deny = ["%s"]`, strings.Join(tableDeny, `", "`))
	}

	return fmt.Sprintf(`
resource "dbtcloud_global_connection" "test_snowflake" {
  name = "%s"

  snowflake = {
    account   = "%s"
    database  = "%s"
    warehouse = "%s"
    allow_sso = false
  }
}

resource "dbtcloud_connection_catalog_config" "test" {
  connection_id = dbtcloud_global_connection.test_snowflake.id

  %s
  %s
  %s
  %s
}
`, connectionName, snowflakeAccount, snowflakeDatabase, snowflakeWarehouse,
		databaseAllowStr, databaseDenyStr, schemaAllowStr, tableDenyStr)
}
