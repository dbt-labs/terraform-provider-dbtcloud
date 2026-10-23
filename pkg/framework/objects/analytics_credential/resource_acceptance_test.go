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
	redshiftHostname, redshiftDatabase                      string
	redshiftUsername, redshiftPassword                      string
}

func accConfig(t *testing.T, needSnowflake, needDatabricks, needPrivateKey bool, needRedshift ...bool) accAnalyticsConfig {
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
	if len(needRedshift) > 0 && needRedshift[0] {
		config.redshiftHostname = os.Getenv("DBT_ACCEPTANCE_TEST_REDSHIFT_HOSTNAME")
		config.redshiftDatabase = os.Getenv("DBT_ACCEPTANCE_TEST_REDSHIFT_DATABASE")
		config.redshiftUsername = os.Getenv("DBT_ACCEPTANCE_TEST_REDSHIFT_USERNAME")
		config.redshiftPassword = os.Getenv("DBT_ACCEPTANCE_TEST_REDSHIFT_PASSWORD")
		if config.redshiftHostname == "" || config.redshiftDatabase == "" ||
			config.redshiftUsername == "" || config.redshiftPassword == "" {
			t.Skip("Skipping: the DBT_ACCEPTANCE_TEST_REDSHIFT_* environment variables are not set.")
		}
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

func (c accAnalyticsConfig) redshiftConnection(name string) string {
	return fmt.Sprintf(`
resource "dbtcloud_global_connection" "redshift" {
  name = "%s"

  redshift = {
    hostname = "%s"
    dbname   = "%s"
  }
}
`, name, c.redshiftHostname, c.redshiftDatabase)
}

func (c accAnalyticsConfig) redshiftCredential(version int, password string) string {
	return fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "redshift" {
  connection_id = dbtcloud_global_connection.redshift.id

  redshift = {
    username            = "%s"
    password_wo         = "%s"
    password_wo_version = %d
  }
}
`, c.redshiftUsername, password, version)
}

// bigqueryConnection creates a BigQuery connection with fixture values: nothing calls GCP
// when a connection or an analytics credential is stored
func bigqueryConnection(name string, latestAdapter bool) string {
	return fmt.Sprintf(`
resource "dbtcloud_global_connection" "bigquery" {
  name = "%s"

  bigquery = {
    gcp_project_id              = "analytics-fixture-project"
    private_key_id              = "fixture-private-key-id"
    private_key                 = "ABCDEFGHIJKL"
    client_email                = "fixture@example.iam.gserviceaccount.com"
    client_id                   = "fixture-client-id"
    auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
    token_uri                   = "https://oauth2.googleapis.com/token"
    auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
    client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/fixture"
    use_latest_adapter          = %t
  }
}
`, name, latestAdapter)
}

func bigqueryServiceAccountCredential(authType string, version int) string {
	authTypeLine := ""
	if authType != "" {
		authTypeLine = fmt.Sprintf("auth_type = %q", authType)
	}
	return fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "bigquery" {
  connection_id = dbtcloud_global_connection.bigquery.id

  bigquery = {
    %s
    private_key_id              = "analytics-private-key-id"
    private_key_wo              = "analytics-fixture-private-key"
    private_key_wo_version      = %d
    client_email                = "analytics@example.iam.gserviceaccount.com"
    client_id                   = "analytics-client-id"
    auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
    token_uri                   = "https://oauth2.googleapis.com/token"
    auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
    client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/analytics"
    execution_project           = "analytics-billing-project"
  }
}
`, authTypeLine, version)
}

const bigqueryWifCredential = `
resource "dbtcloud_analytics_credential" "bigquery" {
  connection_id = dbtcloud_global_connection.bigquery.id

  bigquery = {
    auth_type                   = "external-oauth-wif"
    workload_pool_provider_path = "//iam.googleapis.com/projects/1/locations/global/workloadIdentityPools/fixture/providers/fixture"
  }
}
`

var versionsIgnoredOnImport = []string{
	"bigquery.private_key_wo_version",
	"redshift.password_wo_version",
	"snowflake.password_wo_version",
	"snowflake.private_key_wo_version",
	"snowflake.private_key_passphrase_wo_version",
	"databricks.token_wo_version",
}

// Rule: a Snowflake credential with password authentication is created, rotated and imported without its password reaching the state.
// given: a Snowflake connection and a credential with password_wo
// when: the credential is applied, its version is bumped and it is imported
// then: the state holds no password, the version change sends the secret again and the import matches.
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

// Rule: a Snowflake credential with keypair authentication keeps its private key out of the state.
// given: a Snowflake connection and a credential with private_key_wo
// when: the credential is applied
// then: the state holds no private key.
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

// Rule: a Databricks credential keeps its token out of the state.
// given: a Databricks connection and a credential with token_wo
// when: the credential is applied and imported
// then: the state holds no token and the import matches.
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

// Rule: a Redshift credential keeps its password out of the state.
// given: a Redshift connection and a credential with password_wo
// when: the credential is applied and imported
// then: the state holds no password and the import matches.
func TestAccDbtCloudAnalyticsCredentialRedshift(t *testing.T) {
	config := accConfig(t, false, false, false, true)
	name := randomName()
	address := "dbtcloud_analytics_credential.redshift"

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{
			Config: config.redshiftConnection(name) + config.redshiftCredential(1, config.redshiftPassword),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(address, "adapter_version", "redshift_v0"),
				resource.TestCheckResourceAttr(address, "redshift.username", config.redshiftUsername),
				resource.TestCheckNoResourceAttr(address, "redshift.password_wo"),
				stateHasNoSecret(config.redshiftPassword),
			),
		},
		{
			// the version bump sends the password again
			Config: config.redshiftConnection(name) + config.redshiftCredential(2, config.redshiftPassword),
			Check:  resource.TestCheckResourceAttr(address, "redshift.password_wo_version", "2"),
		},
		{
			Config:                  config.redshiftConnection(name) + config.redshiftCredential(2, config.redshiftPassword),
			ResourceName:            address,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: versionsIgnoredOnImport,
		},
	}))
}

// Rule: a BigQuery service account credential keeps its private key out of the state.
// given: a BigQuery connection and a credential with private_key_wo
// when: the credential is applied and imported
// then: the state holds no private key and the import matches.
func TestAccDbtCloudAnalyticsCredentialBigQueryServiceAccount(t *testing.T) {
	accConfig(t, false, false, false)
	name := randomName()
	address := "dbtcloud_analytics_credential.bigquery"

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{
			Config: bigqueryConnection(name, true) + bigqueryServiceAccountCredential("", 1),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(address, "adapter_version", "bigquery_v1"),
				resource.TestCheckResourceAttr(address, "bigquery.execution_project", "analytics-billing-project"),
				resource.TestCheckNoResourceAttr(address, "bigquery.private_key_wo"),
				stateHasNoSecret("analytics-fixture-private-key"),
			),
		},
		{
			// the version bump sends the private key again
			Config: bigqueryConnection(name, true) + bigqueryServiceAccountCredential("", 2),
			Check:  resource.TestCheckResourceAttr(address, "bigquery.private_key_wo_version", "2"),
		},
		{
			Config:                  bigqueryConnection(name, true) + bigqueryServiceAccountCredential("", 2),
			ResourceName:            address,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: versionsIgnoredOnImport,
		},
	}))
}

// Rule: a BigQuery credential can use workload identity federation without service account keys.
// given: a BigQuery connection and a credential with external-oauth-wif
// when: the credential is applied
// then: it is created with the provider path and no key fields.
func TestAccDbtCloudAnalyticsCredentialBigQueryWorkloadIdentityFederation(t *testing.T) {
	accConfig(t, false, false, false)
	name := randomName()
	address := "dbtcloud_analytics_credential.bigquery"

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{
			Config: bigqueryConnection(name, true) + bigqueryWifCredential,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(address, "adapter_version", "bigquery_v1"),
				resource.TestCheckResourceAttr(address, "bigquery.auth_type", "external-oauth-wif"),
				resource.TestCheckNoResourceAttr(address, "bigquery.private_key_wo"),
			),
		},
		{
			Config:                  bigqueryConnection(name, true) + bigqueryWifCredential,
			ResourceName:            address,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: versionsIgnoredOnImport,
		},
	}))
}

// bigquery_v0 connections are only tested when the account can still create them
func TestAccDbtCloudAnalyticsCredentialBigQueryV0(t *testing.T) {
	accConfig(t, false, false, false)
	if os.Getenv("DBT_ACCEPTANCE_TEST_ANALYTICS_BIGQUERY_V0") != "true" {
		t.Skip("Skipping: set DBT_ACCEPTANCE_TEST_ANALYTICS_BIGQUERY_V0=true when the account can create bigquery_v0 connections.")
	}
	name := randomName()
	address := "dbtcloud_analytics_credential.bigquery"

	resource.Test(t, accTestCase(t, []resource.TestStep{{
		Config: bigqueryConnection(name, false) + bigqueryServiceAccountCredential("", 1),
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(address, "adapter_version", "bigquery_v0"),
			resource.TestCheckNoResourceAttr(address, "bigquery.auth_type"),
			stateHasNoSecret("analytics-fixture-private-key"),
		),
	}}))
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

// Rule: a project can use a connection as its analytics connection, and keeps it on unrelated updates.
// given: a Snowflake connection and a project that uses it as analytics connection
// when: the project is created, updated on another attribute, imported and read by the data source
// then: the analytics connection is set each time, and removing the attribute clears it.
func TestAccDbtCloudAnalyticsConnectionOnProject(t *testing.T) {
	config := accConfig(t, true, false, false)
	name := randomName()
	connection := config.snowflakeConnection(name)
	project := func(attributes string) string {
		return fmt.Sprintf(`
resource "dbtcloud_project" "analytics" {
  name = "%s"
  %s
}

data "dbtcloud_project" "analytics" {
  id = dbtcloud_project.analytics.id
}
`, name, attributes)
	}
	attached := "analytics_connection_id = dbtcloud_global_connection.snowflake.id"

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{
			Config: connection + project(attached),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrPair("dbtcloud_project.analytics", "analytics_connection_id", "dbtcloud_global_connection.snowflake", "id"),
				resource.TestCheckResourceAttrPair("data.dbtcloud_project.analytics", "analytics_connection_id", "dbtcloud_global_connection.snowflake", "id"),
			),
		},
		{
			Config: connection + project(attached+"\n  description = \"updated\""),
			Check:  resource.TestCheckResourceAttrPair("dbtcloud_project.analytics", "analytics_connection_id", "dbtcloud_global_connection.snowflake", "id"),
		},
		{
			Config:            connection + project(attached),
			ResourceName:      "dbtcloud_project.analytics",
			ImportState:       true,
			ImportStateVerify: true,
		},
		{
			Config: connection + project(""),
			Check:  resource.TestCheckNoResourceAttr("dbtcloud_project.analytics", "analytics_connection_id"),
		},
	}))
}

// Rule: a connection shows the analytics credential attached to it.
// given: a Snowflake connection and an analytics credential created on it
// when: the connection is read, as a resource and as a data source
// then: its analytics_credentials_id is the id of the credential.
func TestAccDbtCloudConnectionShowsAnalyticsCredentialsID(t *testing.T) {
	config := accConfig(t, true, false, false)
	name := randomName()

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{
			Config: config.snowflakeConnection(name) + config.snowflakePasswordCredential("snowflake", 1, config.snowflakePassword) + `
data "dbtcloud_global_connection" "snowflake" {
  id = dbtcloud_global_connection.snowflake.id

  depends_on = [dbtcloud_analytics_credential.snowflake]
}
`,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrPair("data.dbtcloud_global_connection.snowflake", "analytics_credentials_id", "dbtcloud_analytics_credential.snowflake", "id"),
			),
		},
		{
			// the attached credential shows on the connection after a refresh
			Config: config.snowflakeConnection(name) + config.snowflakePasswordCredential("snowflake", 1, config.snowflakePassword),
			Check:  resource.TestCheckResourceAttrPair("dbtcloud_global_connection.snowflake", "analytics_credentials_id", "dbtcloud_analytics_credential.snowflake", "id"),
		},
	}))
}
