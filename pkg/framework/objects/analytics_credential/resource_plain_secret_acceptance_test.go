package analytics_credential_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// plainSecretsIgnoredOnImport lists the attributes the API never returns, so an import cannot fill them
var plainSecretsIgnoredOnImport = append([]string{
	"snowflake.password",
	"databricks.token",
	"redshift.password",
	"bigquery.private_key",
}, versionsIgnoredOnImport...)

func (c accAnalyticsConfig) snowflakePlainPasswordCredential(password string) string {
	return fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "snowflake" {
  connection_id = dbtcloud_global_connection.snowflake.id

  snowflake = {
    auth_type = "password"
    user      = "%s"
    role      = "%s"
    warehouse = "%s"
    password  = "%s"
  }
}
`, c.snowflakeUser, c.snowflakeRole, c.snowflakeWarehouse, password)
}

func (c accAnalyticsConfig) databricksPlainTokenCredential(token string) string {
	return fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "databricks" {
  connection_id = dbtcloud_global_connection.databricks.id

  databricks = {
    token   = "%s"
    catalog = "%s"
  }
}
`, token, c.databricksCatalog)
}

func (c accAnalyticsConfig) redshiftPlainPasswordCredential(password string) string {
	return fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "redshift" {
  connection_id = dbtcloud_global_connection.redshift.id

  redshift = {
    username = "%s"
    password = "%s"
  }
}
`, c.redshiftUsername, password)
}

func bigqueryPlainPrivateKeyCredential(privateKey string) string {
	return fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "bigquery" {
  connection_id = dbtcloud_global_connection.bigquery.id

  bigquery = {
    private_key_id              = "analytics-private-key-id"
    private_key                 = "%s"
    client_email                = "analytics@example.iam.gserviceaccount.com"
    client_id                   = "analytics-client-id"
    auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
    token_uri                   = "https://oauth2.googleapis.com/token"
    auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
    client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/analytics"
    execution_project           = "analytics-billing-project"
  }
}
`, privateKey)
}

// Rule: a Snowflake credential can use a plain password, which is stored in the state and replaced when it changes.
// given: a Snowflake connection and a credential with a plain password
// when: the credential is applied, the password is changed and the credential is imported
// then: the state holds the password, the change is applied and the import matches apart from the password.
func TestAccDbtCloudAnalyticsCredentialSnowflakePlainPassword(t *testing.T) {
	config := accConfig(t, true, false, false)
	name := randomName()
	address := "dbtcloud_analytics_credential.snowflake"
	rotated := config.snowflakePassword + "-rotated"

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{
			Config: config.snowflakeConnection(name) + config.snowflakePlainPasswordCredential(config.snowflakePassword),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(address, "adapter_version", "snowflake_v0"),
				resource.TestCheckResourceAttr(address, "snowflake.password", config.snowflakePassword),
			),
		},
		{
			Config: config.snowflakeConnection(name) + config.snowflakePlainPasswordCredential(rotated),
			Check:  resource.TestCheckResourceAttr(address, "snowflake.password", rotated),
		},
		{
			Config:                  config.snowflakeConnection(name) + config.snowflakePlainPasswordCredential(rotated),
			ResourceName:            address,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: plainSecretsIgnoredOnImport,
		},
	}))
}

// Rule: a Databricks credential can use a plain token, which is stored in the state and replaced when it changes.
// given: a Databricks connection and a credential with a plain token
// when: the credential is applied, the token is changed and the credential is imported
// then: the state holds the token, the change is applied and the import matches apart from the token.
func TestAccDbtCloudAnalyticsCredentialDatabricksPlainToken(t *testing.T) {
	config := accConfig(t, false, true, false)
	name := randomName()
	address := "dbtcloud_analytics_credential.databricks"
	rotated := config.databricksToken + "-rotated"

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{
			Config: config.databricksConnection(name) + config.databricksPlainTokenCredential(config.databricksToken),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(address, "adapter_version", "databricks_v0"),
				resource.TestCheckResourceAttr(address, "databricks.token", config.databricksToken),
			),
		},
		{
			Config: config.databricksConnection(name) + config.databricksPlainTokenCredential(rotated),
			Check:  resource.TestCheckResourceAttr(address, "databricks.token", rotated),
		},
		{
			Config:                  config.databricksConnection(name) + config.databricksPlainTokenCredential(rotated),
			ResourceName:            address,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: plainSecretsIgnoredOnImport,
		},
	}))
}

// Rule: a Redshift credential can use a plain password, which is stored in the state and replaced when it changes.
// given: a Redshift connection and a credential with a plain password
// when: the credential is applied, the password is changed and the credential is imported
// then: the state holds the password, the change is applied and the import matches apart from the password.
func TestAccDbtCloudAnalyticsCredentialRedshiftPlainPassword(t *testing.T) {
	config := accConfig(t, false, false, false, true)
	name := randomName()
	address := "dbtcloud_analytics_credential.redshift"
	rotated := config.redshiftPassword + "-rotated"

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{
			Config: config.redshiftConnection(name) + config.redshiftPlainPasswordCredential(config.redshiftPassword),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(address, "adapter_version", "redshift_v0"),
				resource.TestCheckResourceAttr(address, "redshift.password", config.redshiftPassword),
			),
		},
		{
			Config: config.redshiftConnection(name) + config.redshiftPlainPasswordCredential(rotated),
			Check:  resource.TestCheckResourceAttr(address, "redshift.password", rotated),
		},
		{
			Config:                  config.redshiftConnection(name) + config.redshiftPlainPasswordCredential(rotated),
			ResourceName:            address,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: plainSecretsIgnoredOnImport,
		},
	}))
}

// Rule: a BigQuery service account credential can use a plain private key, which is stored in the state and replaced when it changes.
// given: a BigQuery connection and a credential with a plain private key
// when: the credential is applied, the key is changed and the credential is imported
// then: the state holds the key, the change is applied and the import matches apart from the key.
func TestAccDbtCloudAnalyticsCredentialBigQueryPlainPrivateKey(t *testing.T) {
	accConfig(t, false, false, false)
	name := randomName()
	address := "dbtcloud_analytics_credential.bigquery"

	resource.Test(t, accTestCase(t, []resource.TestStep{
		{
			Config: bigqueryConnection(name, true) + bigqueryPlainPrivateKeyCredential("analytics-fixture-private-key"),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(address, "adapter_version", "bigquery_v1"),
				resource.TestCheckResourceAttr(address, "bigquery.private_key", "analytics-fixture-private-key"),
			),
		},
		{
			Config: bigqueryConnection(name, true) + bigqueryPlainPrivateKeyCredential("analytics-fixture-private-key-rotated"),
			Check:  resource.TestCheckResourceAttr(address, "bigquery.private_key", "analytics-fixture-private-key-rotated"),
		},
		{
			Config:            bigqueryConnection(name, true) + bigqueryPlainPrivateKeyCredential("analytics-fixture-private-key-rotated"),
			ResourceName:      address,
			ImportState:       true,
			ImportStateVerify: true,
			// import stores the server auth type, the configuration leaves the default out
			ImportStateVerifyIgnore: append([]string{"bigquery.auth_type"}, plainSecretsIgnoredOnImport...),
		},
	}))
}
