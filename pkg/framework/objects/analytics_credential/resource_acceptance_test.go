package analytics_credential_test

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// These tests run against a real account, so they are skipped unless TF_ACC is set, the
// flag cc-4192-enable-analytics-credentials is on for the account (confirmed with
// DBT_ACCEPTANCE_TEST_ANALYTICS_CREDENTIALS_ENABLED=true) and the warehouse details
// are given through environment variables. They need Terraform >= 1.11.

type accAnalyticsConfig struct {
	snowflakeAccount, snowflakeDatabase, snowflakeWarehouse string
	snowflakeUser, snowflakePassword, snowflakeRole         string
	snowflakePrivateKey                                     string
	databricksHost, databricksHTTPPath                      string
	databricksToken, databricksCatalog                      string
}

func accConfig(t *testing.T, needSnowflake, needDatabricks, needPrivateKey bool) accAnalyticsConfig {
	t.Helper()
	if os.Getenv("DBT_ACCEPTANCE_TEST_ANALYTICS_CREDENTIALS_ENABLED") != "true" {
		t.Skip("Skipping: set DBT_ACCEPTANCE_TEST_ANALYTICS_CREDENTIALS_ENABLED=true when the cc-4192-enable-analytics-credentials flag is on for the account.")
	}

	config := accAnalyticsConfig{}
	if needSnowflake {
		snowflake := acctest_helper.GetPlatformMetadataCredentialTestingConfigurations()
		if snowflake == nil {
			t.Skip("Skipping: the DBT_ACCEPTANCE_TEST_SNOWFLAKE_* environment variables are not set.")
		}
		config.snowflakeAccount = snowflake.SnowflakeAccount
		config.snowflakeDatabase = snowflake.SnowflakeDatabase
		config.snowflakeWarehouse = snowflake.SnowflakeWarehouse
		config.snowflakeUser = snowflake.User
		config.snowflakePassword = snowflake.Password
		config.snowflakeRole = snowflake.Role
	}
	if needPrivateKey {
		config.snowflakePrivateKey = os.Getenv("DBT_ACCEPTANCE_TEST_SNOWFLAKE_PRIVATE_KEY")
		if config.snowflakePrivateKey == "" {
			t.Skip("Skipping: DBT_ACCEPTANCE_TEST_SNOWFLAKE_PRIVATE_KEY is not set.")
		}
	}
	if needDatabricks {
		databricks := acctest_helper.GetDatabricksPlatformMetadataCredentialTestingConfigurations()
		if databricks == nil {
			t.Skip("Skipping: the DBT_ACCEPTANCE_TEST_DATABRICKS_* environment variables are not set.")
		}
		config.databricksHost = databricks.Host
		config.databricksHTTPPath = databricks.HTTPPath
		config.databricksToken = databricks.Token
		config.databricksCatalog = databricks.Catalog
	}
	return config
}

func accTestCase(t *testing.T, steps []resource.TestStep) resource.TestCase {
	return resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAnalyticsCredentialDestroy,
		// write-only attributes require Terraform >= 1.11
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		Steps: steps,
	}
}

func (c accAnalyticsConfig) snowflakeConnection(name string) string {
	return fmt.Sprintf(`
resource "dbtcloud_global_connection" "snowflake" {
  name = "%s"

  snowflake = {
    account   = "%s"
    database  = "%s"
    warehouse = "%s"
    allow_sso = false
  }
}
`, name, c.snowflakeAccount, c.snowflakeDatabase, c.snowflakeWarehouse)
}

func (c accAnalyticsConfig) databricksConnection(name string) string {
	return fmt.Sprintf(`
resource "dbtcloud_global_connection" "databricks" {
  name = "%s"

  databricks = {
    host      = "%s"
    http_path = "%s"
    catalog   = "%s"
  }
}
`, name, c.databricksHost, c.databricksHTTPPath, c.databricksCatalog)
}

func (c accAnalyticsConfig) snowflakePasswordCredential(connection string, version int, password string) string {
	return fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "snowflake" {
  connection_id = dbtcloud_global_connection.%s.id

  snowflake = {
    auth_type           = "password"
    user                = "%s"
    role                = "%s"
    warehouse           = "%s"
    password_wo         = "%s"
    password_wo_version = %d
  }
}
`, connection, c.snowflakeUser, c.snowflakeRole, c.snowflakeWarehouse, password, version)
}

func (c accAnalyticsConfig) snowflakeKeypairCredential(connection string) string {
	return fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "snowflake" {
  connection_id = dbtcloud_global_connection.%s.id

  snowflake = {
    auth_type              = "keypair"
    user                   = "%s"
    warehouse              = "%s"
    private_key_wo         = <<-EOT
%s
EOT
    private_key_wo_version = 1
  }
}
`, connection, c.snowflakeUser, c.snowflakeWarehouse, c.snowflakePrivateKey)
}

func (c accAnalyticsConfig) databricksCredential(connection string, version int, token string) string {
	return fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "databricks" {
  connection_id = dbtcloud_global_connection.%s.id

  databricks = {
    token_wo         = "%s"
    token_wo_version = %d
    catalog          = "%s"
  }
}
`, connection, token, version, c.databricksCatalog)
}

func randomName() string {
	return strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
}

var versionsIgnoredOnImport = []string{
	"snowflake.password_wo_version",
	"snowflake.private_key_wo_version",
	"snowflake.private_key_passphrase_wo_version",
	"databricks.token_wo_version",
}

func TestAccDbtCloudAnalyticsCredentialSnowflakePassword(t *testing.T) {
	config := accConfig(t, true, false, false)
	name := randomName()
	address := "dbtcloud_analytics_credential.snowflake"

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{
			Config: config.snowflakeConnection(name) + config.snowflakePasswordCredential("snowflake", 1, config.snowflakePassword),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet(address, "id"),
				resource.TestCheckResourceAttr(address, "adapter_version", "snowflake_v0"),
				resource.TestCheckResourceAttr(address, "snowflake.auth_type", "password"),
				resource.TestCheckResourceAttr(address, "snowflake.user", config.snowflakeUser),
				resource.TestCheckNoResourceAttr(address, "snowflake.password_wo"),
				stateHasNoSecret(config.snowflakePassword),
			),
		},
		{
			// the version bump sends the password again
			Config: config.snowflakeConnection(name) + config.snowflakePasswordCredential("snowflake", 2, config.snowflakePassword),
			Check:  resource.TestCheckResourceAttr(address, "snowflake.password_wo_version", "2"),
		},
		{
			Config:                  config.snowflakeConnection(name) + config.snowflakePasswordCredential("snowflake", 2, config.snowflakePassword),
			ResourceName:            address,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: versionsIgnoredOnImport,
		},
	}))
}

func TestAccDbtCloudAnalyticsCredentialSnowflakeKeypair(t *testing.T) {
	config := accConfig(t, true, false, true)
	name := randomName()
	address := "dbtcloud_analytics_credential.snowflake"

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{
			Config: config.snowflakeConnection(name) + config.snowflakeKeypairCredential("snowflake"),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(address, "snowflake.auth_type", "keypair"),
				resource.TestCheckNoResourceAttr(address, "snowflake.private_key_wo"),
				stateHasNoSecret(config.snowflakePrivateKey),
			),
		},
		{
			Config:                  config.snowflakeConnection(name) + config.snowflakeKeypairCredential("snowflake"),
			ResourceName:            address,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: versionsIgnoredOnImport,
		},
	}))
}

func TestAccDbtCloudAnalyticsCredentialDatabricks(t *testing.T) {
	config := accConfig(t, false, true, false)
	name := randomName()
	address := "dbtcloud_analytics_credential.databricks"

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{
			Config: config.databricksConnection(name) + config.databricksCredential("databricks", 1, config.databricksToken),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(address, "adapter_version", "databricks_v0"),
				resource.TestCheckResourceAttr(address, "databricks.catalog", config.databricksCatalog),
				resource.TestCheckNoResourceAttr(address, "databricks.token_wo"),
				stateHasNoSecret(config.databricksToken),
			),
		},
		{
			Config: config.databricksConnection(name) + config.databricksCredential("databricks", 2, config.databricksToken),
			Check:  resource.TestCheckResourceAttr(address, "databricks.token_wo_version", "2"),
		},
		{
			Config:                  config.databricksConnection(name) + config.databricksCredential("databricks", 2, config.databricksToken),
			ResourceName:            address,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: versionsIgnoredOnImport,
		},
	}))
}

// Rule: the adapter of a credential cannot change in place.
// given: a snowflake credential
// when: the block is swapped for databricks together with a Databricks connection, or on the same connection
// then: the first is a destroy-before-create replace and the second fails at plan time
func TestAccDbtCloudAnalyticsCredentialAdapterSwap(t *testing.T) {
	config := accConfig(t, true, true, false)
	name := randomName()
	address := "dbtcloud_analytics_credential.snowflake"
	swappedAddress := "dbtcloud_analytics_credential.databricks"

	connections := config.snowflakeConnection(name) + config.databricksConnection(name+"DBX")

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{Config: connections + config.snowflakePasswordCredential("snowflake", 1, config.snowflakePassword)},
		{
			// swapping the block on the same connection never reaches the destroy step
			Config: connections + fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "snowflake" {
  connection_id = dbtcloud_global_connection.snowflake.id

  databricks = {
    token_wo         = "%s"
    token_wo_version = 1
  }
}
`, config.databricksToken),
			ExpectError: regexp.MustCompile(`Adapter mismatch`),
		},
		{
			Config: connections + config.databricksCredential("databricks", 1, config.databricksToken),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionDestroy),
					plancheck.ExpectResourceAction(swappedAddress, plancheck.ResourceActionCreate),
				},
			},
		},
	}))
}

// Rule: the adapter of a credential cannot change in place.
// given: a snowflake credential whose block is replaced by a databricks block under the same address
// when: the connection moves to a Databricks connection
// then: the plan is a destroy-before-create replace
func TestAccDbtCloudAnalyticsCredentialSameAddressSwapReplaces(t *testing.T) {
	config := accConfig(t, true, true, false)
	name := randomName()
	address := "dbtcloud_analytics_credential.snowflake"

	connections := config.snowflakeConnection(name) + config.databricksConnection(name+"DBX")

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{Config: connections + config.snowflakePasswordCredential("snowflake", 1, config.snowflakePassword)},
		{
			Config: connections + fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "snowflake" {
  connection_id = dbtcloud_global_connection.databricks.id

  databricks = {
    token_wo         = "%s"
    token_wo_version = 1
  }
}
`, config.databricksToken),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionDestroyBeforeCreate),
				},
			},
		},
	}))
}

// Rule: secrets never reach the state.
// given: a credential created with write-only secrets
// when: it is applied
// then: the write-only attributes are null in the state and a scan of the state finds no secret value
func TestAccDbtCloudAnalyticsCredentialSecretsNeverInState(t *testing.T) {
	config := accConfig(t, true, false, false)
	name := randomName()
	address := "dbtcloud_analytics_credential.snowflake"

	resource.Test(t, accTestCase(t, []resource.TestStep{{
		Config: config.snowflakeConnection(name) + config.snowflakePasswordCredential("snowflake", 1, config.snowflakePassword),
		ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue(address, tfjsonpath.New("snowflake").AtMapKey("password_wo"), knownvalue.Null()),
			statecheck.ExpectKnownValue(address, tfjsonpath.New("snowflake").AtMapKey("private_key_wo"), knownvalue.Null()),
		},
		Check: stateHasNoSecret(config.snowflakePassword),
	}}))
}

// Rule: a credential has exactly one adapter block.
// given: a credential with a snowflake and a databricks block
// when: the configuration is validated
// then: it fails with the expected error and nothing is applied
func TestAccDbtCloudAnalyticsCredentialTwoBlocksFail(t *testing.T) {
	config := accConfig(t, true, true, false)
	name := randomName()

	resource.Test(t, accTestCase(t, []resource.TestStep{{
		Config: config.snowflakeConnection(name) + fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "snowflake" {
  connection_id = dbtcloud_global_connection.snowflake.id

  snowflake = {
    auth_type           = "password"
    user                = "%s"
    warehouse           = "%s"
    password_wo         = "%s"
    password_wo_version = 1
  }

  databricks = {
    token_wo         = "%s"
    token_wo_version = 1
  }
}
`, config.snowflakeUser, config.snowflakeWarehouse, config.snowflakePassword, config.databricksToken),
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
	}}))
}

func testAccCheckAnalyticsCredentialDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("issue getting the client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_analytics_credential" {
			continue
		}

		id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert the id to int64: %w", err)
		}

		_, err = apiClient.GetAnalyticsCredential(id)
		if err == nil {
			return fmt.Errorf("analytics credential still exists")
		}
		if !strings.Contains(err.Error(), "resource-not-found") {
			return fmt.Errorf("expected resource-not-found, got %s", err)
		}
	}

	return nil
}
