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

func TestAccDbtCloudSnowflakePlatformMetadataCredentialResourceWriteOnly(t *testing.T) {
	t.Skip("Skipping write-only acceptance test until CI environment supports it")

	config := acctest_helper.GetPlatformMetadataCredentialTestingConfigurations()
	if config == nil {
		t.Skip("Skipping test because required environment variables are not set.")
	}

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudSnowflakePlatformMetadataCredentialDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with write-only password
			{
				Config: testAccDbtCloudSnowflakePlatformMetadataCredentialWriteOnlyConfig(
					connectionName,
					config.SnowflakeAccount,
					config.SnowflakeDatabase,
					config.SnowflakeWarehouse,
					config.User,
					config.Password,
					config.Role,
					1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("dbtcloud_snowflake_platform_metadata_credential.test_wo", "id"),
					resource.TestCheckResourceAttrSet("dbtcloud_snowflake_platform_metadata_credential.test_wo", "credential_id"),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test_wo", "user", config.User),
					// password_wo should not be in state
					resource.TestCheckNoResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test_wo", "password_wo"),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test_wo", "password_wo_version", "1"),
				),
			},
			// Step 2: Update by incrementing version with new password
			{
				Config: testAccDbtCloudSnowflakePlatformMetadataCredentialWriteOnlyConfig(
					connectionName,
					config.SnowflakeAccount,
					config.SnowflakeDatabase,
					config.SnowflakeWarehouse,
					config.User,
					"new_password",
					config.Role,
					2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test_wo", "password_wo"),
					resource.TestCheckResourceAttr("dbtcloud_snowflake_platform_metadata_credential.test_wo", "password_wo_version", "2"),
				),
			},
			// Step 3: Import
			{
				ResourceName:      "dbtcloud_snowflake_platform_metadata_credential.test_wo",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password", "private_key", "private_key_passphrase",
					"password_wo", "password_wo_version",
					"private_key_wo", "private_key_wo_version",
					"private_key_passphrase_wo", "private_key_passphrase_wo_version",
				},
			},
		},
	})
}

func testAccDbtCloudSnowflakePlatformMetadataCredentialWriteOnlyConfig(
	connectionName string,
	snowflakeAccount string,
	snowflakeDatabase string,
	snowflakeWarehouse string,
	user string,
	passwordWo string,
	role string,
	passwordWoVersion int,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_global_connection" "test_snowflake_wo" {
  name = "%s"

  snowflake = {
    account   = "%s"
    database  = "%s"
    warehouse = "%s"
    allow_sso = false
  }
}

resource "dbtcloud_snowflake_platform_metadata_credential" "test_wo" {
  connection_id = dbtcloud_global_connection.test_snowflake_wo.id

  catalog_ingestion_enabled = true
  cost_optimization_enabled = false
  cost_insights_enabled     = false

  auth_type           = "password"
  user                = "%s"
  password_wo         = "%s"
  password_wo_version = %d
  role                = "%s"
  warehouse           = "%s"
}
`, connectionName, snowflakeAccount, snowflakeDatabase, snowflakeWarehouse,
		user, passwordWo, passwordWoVersion, role, snowflakeWarehouse)
}

func TestAccDbtCloudDatabricksPlatformMetadataCredentialResourceWriteOnly(t *testing.T) {
	t.Skip("Skipping write-only acceptance test until CI environment supports it")

	config := acctest_helper.GetDatabricksPlatformMetadataCredentialTestingConfigurations()
	if config == nil {
		t.Skip("Skipping test because required environment variables are not set.")
	}

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudDatabricksPlatformMetadataCredentialDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with write-only token
			{
				Config: testAccDbtCloudDatabricksPlatformMetadataCredentialWriteOnlyConfig(
					connectionName,
					config.Host,
					config.HTTPPath,
					config.Token,
					config.Catalog,
					1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("dbtcloud_databricks_platform_metadata_credential.test_wo", "id"),
					resource.TestCheckResourceAttrSet("dbtcloud_databricks_platform_metadata_credential.test_wo", "credential_id"),
					resource.TestCheckResourceAttr("dbtcloud_databricks_platform_metadata_credential.test_wo", "catalog", config.Catalog),
					// token_wo should not be in state
					resource.TestCheckNoResourceAttr("dbtcloud_databricks_platform_metadata_credential.test_wo", "token_wo"),
					resource.TestCheckResourceAttr("dbtcloud_databricks_platform_metadata_credential.test_wo", "token_wo_version", "1"),
				),
			},
			// Step 2: Update by incrementing version with new token
			{
				Config: testAccDbtCloudDatabricksPlatformMetadataCredentialWriteOnlyConfig(
					connectionName,
					config.Host,
					config.HTTPPath,
					"new_token_value",
					config.Catalog,
					2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("dbtcloud_databricks_platform_metadata_credential.test_wo", "token_wo"),
					resource.TestCheckResourceAttr("dbtcloud_databricks_platform_metadata_credential.test_wo", "token_wo_version", "2"),
				),
			},
			// Step 3: Import
			{
				ResourceName:      "dbtcloud_databricks_platform_metadata_credential.test_wo",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"token",
					"token_wo", "token_wo_version",
				},
			},
		},
	})
}

func testAccDbtCloudDatabricksPlatformMetadataCredentialWriteOnlyConfig(
	connectionName string,
	host string,
	httpPath string,
	tokenWo string,
	catalog string,
	tokenWoVersion int,
) string {
	return fmt.Sprintf(`
resource "dbtcloud_global_connection" "test_databricks_wo" {
  name = "%s"

  databricks = {
    host      = "%s"
    http_path = "%s"
    catalog   = "%s"
  }
}

resource "dbtcloud_databricks_platform_metadata_credential" "test_wo" {
  connection_id = dbtcloud_global_connection.test_databricks_wo.id

  catalog_ingestion_enabled = true
  cost_optimization_enabled = false
  cost_insights_enabled     = false

  token_wo         = "%s"
  token_wo_version = %d
  catalog          = "%s"
}
`, connectionName, host, httpPath, catalog, tokenWo, tokenWoVersion, catalog)
}

func TestAccDbtCloudDatabricksPlatformMetadataCredentialResource(t *testing.T) {
	config := acctest_helper.GetDatabricksPlatformMetadataCredentialTestingConfigurations()
	if config == nil {
		t.Skip("Skipping test because required environment variables are not set. " +
			"Set DBT_ACCEPTANCE_TEST_DATABRICKS_HOST, DBT_ACCEPTANCE_TEST_DATABRICKS_HTTP_PATH, DBT_ACCEPTANCE_TEST_DATABRICKS_TOKEN, and DBT_ACCEPTANCE_TEST_DATABRICKS_CATALOG to run this test.")
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
					config.Host,
					config.HTTPPath,
					config.Token,
					config.Catalog,
					true,  // catalog_ingestion_enabled
					false, // cost_optimization_enabled
					false, // cost_insights_enabled
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("dbtcloud_databricks_platform_metadata_credential.test", "id"),
					resource.TestCheckResourceAttrSet("dbtcloud_databricks_platform_metadata_credential.test", "credential_id"),
					resource.TestCheckResourceAttr("dbtcloud_databricks_platform_metadata_credential.test", "catalog_ingestion_enabled", "true"),
					resource.TestCheckResourceAttr("dbtcloud_databricks_platform_metadata_credential.test", "catalog", config.Catalog),
				),
			},
			// Update
			{
				Config: testAccDbtCloudDatabricksPlatformMetadataCredentialResourceConfig(
					connectionName,
					config.Host,
					config.HTTPPath,
					config.Token,
					config.Catalog,
					true, // catalog_ingestion_enabled
					true, // cost_optimization_enabled - changed
					true, // cost_insights_enabled - changed
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("dbtcloud_databricks_platform_metadata_credential.test", "catalog_ingestion_enabled", "true"),
					resource.TestCheckResourceAttr("dbtcloud_databricks_platform_metadata_credential.test", "cost_optimization_enabled", "true"),
					resource.TestCheckResourceAttr("dbtcloud_databricks_platform_metadata_credential.test", "cost_insights_enabled", "true"),
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
	host string,
	httpPath string,
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
    host      = "%s"
    http_path = "%s"
    catalog   = "%s"
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
`, connectionName, host, httpPath, catalog, catalogIngestionEnabled, costOptimizationEnabled, costInsightsEnabled, token, catalog)
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
