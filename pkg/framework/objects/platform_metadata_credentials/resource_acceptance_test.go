package platform_metadata_credentials_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudSnowflakePlatformMetadataCredentialResource(t *testing.T) {
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
		CheckDestroy:             testAccCheckDbtCloudSnowflakePlatformMetadataCredentialDestroy,
		Steps: []resource.TestStep{
			// Create with Snowflake password auth - catalog ingestion only
			{
				Config: testAccDbtCloudSnowflakePlatformMetadataCredentialResourceConfig(
					connectionName,
					config.SnowflakeAccount,
					config.SnowflakeDatabase,
					config.SnowflakeWarehouse,
					config.User,
					config.Password,
					config.Role,
					true,  // catalog_ingestion_enabled
					false, // cost_optimization_enabled
					false, // cost_insights_enabled
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("dbtcloud_snowflake_platform_metadata_credential.test", "id"),
					resource.TestCheckResourceAttrSet("dbtcloud_snowflake_platform_metadata_credential.test", "credential_id"),
					resource.TestCheckResourceAttrSet("dbtcloud_snowflake_platform_metadata_credential.test", "connection_id"),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "catalog_ingestion_enabled", "true"),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "cost_optimization_enabled", "false"),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "cost_insights_enabled", "false"),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "auth_type", "password"),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "user", config.User),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "role", config.Role),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "warehouse", config.SnowflakeWarehouse),
					resource.TestCheckResourceAttrSet("dbtcloud_snowflake_platform_metadata_credential.test", "adapter_version"),
				),
			},
			// Update feature flags - enable all
			{
				Config: testAccDbtCloudSnowflakePlatformMetadataCredentialResourceConfig(
					connectionName,
					config.SnowflakeAccount,
					config.SnowflakeDatabase,
					config.SnowflakeWarehouse,
					config.User,
					config.Password,
					config.Role,
					true, // catalog_ingestion_enabled
					true, // cost_optimization_enabled
					true, // cost_insights_enabled
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "catalog_ingestion_enabled", "true"),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "cost_optimization_enabled", "true"),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "cost_insights_enabled", "true"),
				),
			},
			// Update feature flags - disable some
			{
				Config: testAccDbtCloudSnowflakePlatformMetadataCredentialResourceConfig(
					connectionName,
					config.SnowflakeAccount,
					config.SnowflakeDatabase,
					config.SnowflakeWarehouse,
					config.User,
					config.Password,
					config.Role,
					true,  // catalog_ingestion_enabled
					false, // cost_optimization_enabled
					true,  // cost_insights_enabled
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "catalog_ingestion_enabled", "true"),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "cost_optimization_enabled", "false"),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test", "cost_insights_enabled", "true"),
				),
			},
			// Import test
			{
				ResourceName:      "dbtcloud_snowflake_platform_metadata_credential.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Sensitive fields won't match after import because the API returns masked values
				ImportStateVerifyIgnore: []string{
					"password",
					"private_key",
					"private_key_passphrase",
				},
			},
		},
	})
}

func testAccDbtCloudSnowflakePlatformMetadataCredentialResourceConfig(
	connectionName string,
	snowflakeAccount string,
	snowflakeDatabase string,
	snowflakeWarehouse string,
	user string,
	password string,
	role string,
	catalogIngestionEnabled bool,
	costOptimizationEnabled bool,
	costInsightsEnabled bool,
) string {
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

resource "dbtcloud_snowflake_platform_metadata_credential" "test" {
  connection_id = dbtcloud_global_connection.test_snowflake.id

  catalog_ingestion_enabled = %t
  cost_optimization_enabled = %t
  cost_insights_enabled     = %t

  auth_type = "password"
  user      = "%s"
  password  = "%s"
  role      = "%s"
  warehouse = "%s"
}
`, connectionName, snowflakeAccount, snowflakeDatabase, snowflakeWarehouse,
		catalogIngestionEnabled, costOptimizationEnabled, costInsightsEnabled,
		user, password, role, snowflakeWarehouse)
}

func testAccCheckDbtCloudSnowflakePlatformMetadataCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("issue getting the client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_snowflake_platform_metadata_credential" {
			continue
		}

		credentialIDStr := rs.Primary.Attributes["credential_id"]
		credentialID, err := strconv.ParseInt(credentialIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert credential_id to int64: %w", err)
		}

		_, err = apiClient.GetPlatformMetadataCredential(credentialID)
		if err == nil {
			return fmt.Errorf("Snowflake platform metadata credential still exists")
		}

		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}

func TestAccDbtCloudDatabricksPlatformMetadataCredentialResource(t *testing.T) {
	config := acctest_helper.GetPlatformMetadataCredentialTestingConfigurations()
	if config == nil {
		t.Skip("Skipping test because required environment variables are not set. " +
			"Set ACC_TEST_DATABRICKS_HOST, ACC_TEST_DATABRICKS_HTTP_PATH, ACC_TEST_DATABRICKS_TOKEN, and ACC_TEST_DATABRICKS_CATALOG to run this test.")
	}

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudDatabricksPlatformMetadataCredentialDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccDbtCloudDatabricksPlatformMetadataCredentialResourceConfig(
					connectionName,
					"test_token",
					"main",
					true,  // catalog_ingestion_enabled
					false, // cost_optimization_enabled
					false, // cost_insights_enabled
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("dbtcloud_databricks_platform_metadata_credential.test", "id"),
					resource.TestCheckResourceAttrSet("dbtcloud_databricks_platform_metadata_credential.test", "credential_id"),
					resource.TestCheckResourceAttr("dbtcloud_databricks_platform_metadata_credential.test", "catalog_ingestion_enabled", "true"),
					resource.TestCheckResourceAttr("dbtcloud_databricks_platform_metadata_credential.test", "catalog", "main"),
				),
			},
			// Import test
			{
				ResourceName:      "dbtcloud_databricks_platform_metadata_credential.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"token",
				},
			},
		},
	})
}

func testAccDbtCloudDatabricksPlatformMetadataCredentialResourceConfig(
	connectionName string,
	token string,
	catalog string,
	catalogIngestionEnabled bool,
	costOptimizationEnabled bool,
	costInsightsEnabled bool,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_global_connection" "test_databricks" {
  name = "%s"

  databricks = {
    host      = "test.cloud.databricks.com"
    http_path = "/sql/1.0/warehouses/abc123"
  }
}

resource "dbtcloud_databricks_platform_metadata_credential" "test" {
  connection_id = dbtcloud_global_connection.test_databricks.id

  catalog_ingestion_enabled = %t
  cost_optimization_enabled = %t
  cost_insights_enabled     = %t

  token   = "%s"
  catalog = "%s"
}
`, connectionName, catalogIngestionEnabled, costOptimizationEnabled, costInsightsEnabled, token, catalog)
}

func testAccCheckDbtCloudDatabricksPlatformMetadataCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("issue getting the client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_databricks_platform_metadata_credential" {
			continue
		}

		credentialIDStr := rs.Primary.Attributes["credential_id"]
		credentialID, err := strconv.ParseInt(credentialIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert credential_id to int64: %w", err)
		}

		_, err = apiClient.GetPlatformMetadataCredential(credentialID)
		if err == nil {
			return fmt.Errorf("Databricks platform metadata credential still exists")
		}

		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
