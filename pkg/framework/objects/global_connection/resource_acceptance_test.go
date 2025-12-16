package global_connection_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudGlobalConnectionSnowflakeResource(t *testing.T) {
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	oAuthClientID := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	oAuthClientSecret := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionSnowflakeResourceBasicConfig(
					connectionName,
					oAuthClientID,
					oAuthClientSecret,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"snowflake_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
				),
			},
			// modify, adding optional fields
			{
				Config: testAccDbtCloudSGlobalConnectionSnowflakeResourceFullConfig(
					connectionName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"snowflake_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"oauth_configuration_id",
					),
				),
			},
			// modify, removing optional fields to check PATCH when we remove fields
			{
				Config: testAccDbtCloudSGlobalConnectionSnowflakeResourceBasicConfig(
					connectionName2,
					oAuthClientID,
					oAuthClientSecret,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"snowflake_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_global_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"snowflake.oauth_client_id",
					"snowflake.oauth_client_secret",
				},
			},
		},
	})

}

func testAccDbtCloudSGlobalConnectionSnowflakeResourceBasicConfig(
	connectionName, oAuthClientID, oAuthClientSecret string,
) string {
	return fmt.Sprintf(`

resource dbtcloud_oauth_configuration test {
  type = "entra"
  name = "OAuth config"
  client_secret = "secret"
  client_id = "myid2"
  redirect_uri = "http://example.com"
  token_url = "http://example.com"
  authorize_url = "http://example.com"
  application_id_uri = "app-uri"
}


resource dbtcloud_global_connection test {
  name = "%s"

  snowflake = {
    account = "account"
    warehouse = "warehouse"
    database = "database"
    allow_sso = true
    oauth_client_id = "%s"
    oauth_client_secret = "%s"
    client_session_keep_alive = false
  }
}

`, connectionName, oAuthClientID, oAuthClientSecret)
}

func testAccDbtCloudSGlobalConnectionSnowflakeResourceFullConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`

resource dbtcloud_oauth_configuration test {
  type = "entra"
  name = "OAuth config"
  client_secret = "secret"
  client_id = "myid2"
  redirect_uri = "http://example.com"
  token_url = "http://example.com"
  authorize_url = "http://example.com"
  application_id_uri = "app-uri"
}


resource dbtcloud_global_connection test {
  name = "%s"
  oauth_configuration_id = dbtcloud_oauth_configuration.test.id

  snowflake = {
    account = "account"
    warehouse = "warehouse"
    database = "database"
    allow_sso = false
    client_session_keep_alive = false

	// optional fields
	role = "role"
  }
}
`, connectionName)
}

func TestAccDbtCloudGlobalConnectionBigQueryResource(t *testing.T) {
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	timeoutSeconds := int64(1000)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// // create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionBigQueryResourceBasicConfig(
					connectionName,
					timeoutSeconds,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"bigquery_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
				),
			},
			// modify, adding optional fields
			{
				Config: testAccDbtCloudSGlobalConnectionBigQueryResourceFullConfig(
					connectionName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"bigquery_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"bigquery.deployment_env_auth_type",
						"service-account-json",
					),
				),
			},
			// modify, removing optional fields to check PATCH when we remove fields
			{
				Config: testAccDbtCloudSGlobalConnectionBigQueryResourceBasicConfig(
					connectionName2,
					timeoutSeconds,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"bigquery_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_global_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bigquery.private_key",
					"bigquery.application_secret",
					"bigquery.application_id",
				},
			},
		},
	})

}

func TestAccDbtCloudGlobalConnectionBigQueryCreateV1Adapter(t *testing.T) {
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	jobExecutionTimeoutSeconds := int64(1000)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionBigQueryResourceBasicConfigWithJobExecutionTimeoutSeconds(
					connectionName,
					jobExecutionTimeoutSeconds,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"bigquery_v1",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"bigquery.job_execution_timeout_seconds",
						"1000",
					),
				),
			},

			// IMPORT
			{
				ResourceName:      "dbtcloud_global_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bigquery.private_key",
					"bigquery.application_secret",
					"bigquery.application_id",
					"bigquery.timeout_seconds",
					"bigquery.job_execution_timeout_seconds",
					"bigquery.use_latest_adapter",
				},
			},
		},
	})

}

func TestAccDbtCloudGlobalConnectionBigQueryExternalOAuthWIF(t *testing.T) {
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with external-oauth-wif auth type
			{
				Config: testAccDbtCloudSGlobalConnectionBigQueryResourceExternalOAuthWIFConfig(
					connectionName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"bigquery.deployment_env_auth_type",
						"external-oauth-wif",
					),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_global_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bigquery.application_secret",
					"bigquery.application_id",
				},
			},
		},
	})
}

func testAccDbtCloudSGlobalConnectionBigQueryResourceExternalOAuthWIFConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  bigquery = {
    gcp_project_id         = "my-gcp-project-id"
    application_id         = "oauth_application_id"
    application_secret     = "oauth_secret_id"
    deployment_env_auth_type = "external-oauth-wif"
  }
}

`, connectionName)
}

func TestAccDbtCloudGlobalConnectionBigQueryUpdateV1AdapterFromV0(t *testing.T) {
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	jobExecutionTimeoutSeconds := int64(1000)
	timeoutSeconds := int64(500)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionBigQueryResourceBasicConfigWithTimeoutSeconds(
					connectionName,
					timeoutSeconds,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"bigquery_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"bigquery.timeout_seconds",
						"500",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
				),
			},
			// modify, adding optional fields
			{
				Config: testAccDbtCloudSGlobalConnectionBigQueryResourceBasicConfigWithJobExecutionTimeoutSeconds(
					connectionName,
					jobExecutionTimeoutSeconds,
				),
				// expect an error
				ExpectError: regexp.MustCompile("Changing the adapter version is not supported."),
			},

			// IMPORT
			{
				ResourceName:      "dbtcloud_global_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bigquery.private_key",
					"bigquery.application_secret",
					"bigquery.application_id",
					"bigquery.timeout_seconds",
					"bigquery.adapter_version_override",
				},
			},
		},
	})

}

func testAccDbtCloudSGlobalConnectionBigQueryResourceBasicConfig(
	connectionName string,
	timeoutSeconds int64,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  bigquery = {

    gcp_project_id              = "my-gcp-project-id"
    timeout_seconds             = %d
    private_key_id              = "my-private-key-id"
    private_key                 = "ABCDEFGHIJKL"
    client_email                = "my_client_email"
    client_id                   = "my_client_id"
    auth_uri                    = "my_auth_uri"
    token_uri                   = "my_token_uri"
    auth_provider_x509_cert_url = "my_auth_provider_x509_cert_url"
    client_x509_cert_url        = "my_client_x509_cert_url"
    application_id              = "oauth_application_id"
    application_secret          = "oauth_secret_id"

  }
}

`, connectionName, timeoutSeconds)
}

func testAccDbtCloudSGlobalConnectionBigQueryResourceBasicConfigWithJobExecutionTimeoutSeconds(
	connectionName string,
	jobExecutionTimeoutSeconds int64,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  bigquery = {

    gcp_project_id              = "70403103977025"
    private_key_id              = "my-private-key-id"
    private_key                 = "ABCDEFGHIJKL"
    client_email                = "my_client_email"
    client_id                   = "my_client_id"
    auth_uri                    = "my_auth_uri"
    token_uri                   = "my_token_uri"
    auth_provider_x509_cert_url = "my_auth_provider_x509_cert_url"
    client_x509_cert_url        = "my_client_x509_cert_url"
    application_id              = "oauth_application_id"
    application_secret          = "oauth_secret_id"
	job_execution_timeout_seconds = %d
	use_latest_adapter = true
  }
}

`, connectionName, jobExecutionTimeoutSeconds)
}

func testAccDbtCloudSGlobalConnectionBigQueryResourceBasicConfigWithTimeoutSeconds(
	connectionName string,
	timeoutSeconds int64,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  bigquery = {

    gcp_project_id              = "70403103977025"
    private_key_id              = "my-private-key-id"
    private_key                 = "ABCDEFGHIJKL"
    client_email                = "my_client_email"
    client_id                   = "my_client_id"
    auth_uri                    = "my_auth_uri"
    token_uri                   = "my_token_uri"
    auth_provider_x509_cert_url = "my_auth_provider_x509_cert_url"
    client_x509_cert_url        = "my_client_x509_cert_url"
    application_id              = "oauth_application_id"
    application_secret          = "oauth_secret_id"
	timeout_seconds = %d
	use_legacy_adapter = true
  }
}

`, connectionName, timeoutSeconds)
}

func testAccDbtCloudSGlobalConnectionBigQueryResourceFullConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  bigquery = {

    gcp_project_id              = "my-gcp-project-id"
    timeout_seconds             = 1000
    private_key_id              = "my-private-key-id"
    private_key                 = "ABCDEFGHIJKL"
    client_email                = "my_client_email"
    client_id                   = "my_client_id"
    auth_uri                    = "my_auth_uri"
    token_uri                   = "my_token_uri"
    auth_provider_x509_cert_url = "my_auth_provider_x509_cert_url"
    client_x509_cert_url        = "my_client_x509_cert_url"
    application_id              = "oauth_application_id"
    application_secret          = "oauth_secret_id"
    timeout_seconds 			= 1000

    dataproc_cluster_name        = "dataproc"
    dataproc_region              = "region"
    execution_project            = "project"
    gcs_bucket                   = "bucket"
    impersonate_service_account  = "service"
    job_creation_timeout_seconds = 1000
    job_retry_deadline_seconds   = 1000
    location                     = "us"
    maximum_bytes_billed         = 1000
    priority                     = "batch"
    retries                      = 3
    scopes                       = ["dummyscope"]
    deployment_env_auth_type     = "service-account-json"

  }
}
`, connectionName)
}

func TestAccDbtCloudGlobalConnectionDatabricksResource(t *testing.T) {
	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	oAuthClientID := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	oAuthClientSecret := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionDatabricksResourceBasicConfig(
					connectionName,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"databricks_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
				),
			},
			// modify, adding optional fields
			{
				Config: testAccDbtCloudSGlobalConnectionDatabricksResourceFullConfig(
					connectionName,
					oAuthClientID,
					oAuthClientSecret,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"databricks_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
				),
			},
			// modify, removing optional fields to check PATCH when we remove fields
			{
				Config: testAccDbtCloudSGlobalConnectionDatabricksResourceBasicConfig(
					connectionName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"databricks_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_global_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"databricks.client_id",
					"databricks.client_secret",
				},
			},
		},
	})

}

func testAccDbtCloudSGlobalConnectionDatabricksResourceBasicConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  databricks = {
    host = "databricks.com"
    http_path = "/sql/your/http/path"

	// optional fields
	catalog = "dbt_catalog"
  }
}

`, connectionName)
}

func testAccDbtCloudSGlobalConnectionDatabricksResourceFullConfig(
	connectionName, oAuthClientID, oAuthClientSecret string,
) string {
	return fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  databricks = {
    host = "databricks.com"
    http_path = "/sql/your/http/path"

	// optional fields
	// catalog = "dbt_catalog"
	client_id = "%s"
	client_secret = "%s"
  }
}
`, connectionName, oAuthClientID, oAuthClientSecret)
}

func TestAccDbtCloudGlobalConnectionRedshiftResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionRedshiftResourceBasicConfig(
					connectionName,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"redshift_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
				),
			},
			// modify, adding optional fields
			{
				Config: testAccDbtCloudSGlobalConnectionRedshiftResourceFullConfig(
					connectionName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"redshift_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"true",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"redshift.ssh_tunnel.public_key",
					),
				),
			},
			// IMPORT WITH ALL FIELDS
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// modify, removing optional fields to check PATCH when we remove fields
			{
				Config: testAccDbtCloudSGlobalConnectionRedshiftResourceBasicConfig(
					connectionName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"redshift_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
				),
			},
			// IMPORT SUBSET
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})

}

func testAccDbtCloudSGlobalConnectionRedshiftResourceBasicConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  redshift = {
    hostname = "test.com"
	port = 9876
	dbname = "my_database"
  }
}

`, connectionName)
}

func testAccDbtCloudSGlobalConnectionRedshiftResourceFullConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  redshift = {
    hostname = "test.com"
	port = 1234
	dbname = "my_database"

	// optional fields
    ssh_tunnel = {
      username = "user"
      hostname = "host2"
      port = 1110
    }

  }
}
`, connectionName)
}

func TestAccDbtCloudGlobalConnectionPostgresResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionPostgresResourceBasicConfig(
					connectionName,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"postgres_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
				),
			},
			// modify, adding optional fields
			{
				Config: testAccDbtCloudSGlobalConnectionPostgresResourceFullConfig(
					connectionName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"postgres_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"true",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"postgres.ssh_tunnel.public_key",
					),
				),
			},
			// IMPORT WITH ALL FIELDS
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// modify, removing optional fields to check PATCH when we remove fields
			{
				Config: testAccDbtCloudSGlobalConnectionPostgresResourceBasicConfig(
					connectionName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"postgres_v0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"is_ssh_tunnel_enabled",
						"false",
					),
				),
			},
			// IMPORT SUBSET
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})

}

func testAccDbtCloudSGlobalConnectionPostgresResourceBasicConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  postgres = {
    hostname = "test.com"
	port = 9876
	dbname = "my_database"
  }
}

`, connectionName)
}

func testAccDbtCloudSGlobalConnectionPostgresResourceFullConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  postgres = {
    hostname = "test.com"
	port = 1234
	dbname = "my_database"

	// optional fields
    ssh_tunnel = {
      username = "user"
      hostname = "host2"
      port = 1110
    }

  }
}
`, connectionName)
}

func TestAccDbtCloudGlobalConnectionFabricResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionFabricResourceBasicConfig(
					connectionName,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"fabric_v0",
					),
				),
			},
			// modify, adding optional fields
			{
				Config: testAccDbtCloudSGlobalConnectionFabricResourceFullConfig(
					connectionName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"fabric_v0",
					),
				),
			},
			// IMPORT WITH ALL FIELDS
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// modify, removing optional fields to check PATCH when we remove fields
			{
				Config: testAccDbtCloudSGlobalConnectionFabricResourceBasicConfig(
					connectionName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"fabric_v0",
					),
				),
			},
			// IMPORT SUBSET
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})

}

func testAccDbtCloudSGlobalConnectionFabricResourceBasicConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  fabric = {
    server = "fabric.com"
    database = "fabric"
  }
}

`, connectionName)
}

func testAccDbtCloudSGlobalConnectionFabricResourceFullConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  fabric = {
    server = "fabric.com"
    database = "fabric"

	// optional fields
	port = 1234
	retries = 3
	login_timeout = 1000
	query_timeout = 3600
  }
}
`, connectionName)
}

func TestAccDbtCloudGlobalConnectionSynapseResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionSynapseResourceBasicConfig(
					connectionName,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"synapse_v0",
					),
				),
			},
			// modify, adding optional fields
			{
				Config: testAccDbtCloudSGlobalConnectionSynapseResourceFullConfig(
					connectionName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"synapse_v0",
					),
				),
			},
			// IMPORT WITH ALL FIELDS
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// modify, removing optional fields to check PATCH when we remove fields
			{
				Config: testAccDbtCloudSGlobalConnectionSynapseResourceBasicConfig(
					connectionName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"synapse_v0",
					),
				),
			},
			// IMPORT SUBSET
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})

}

func testAccDbtCloudSGlobalConnectionSynapseResourceBasicConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  synapse = {
    host = "synapse.com"
    database = "synapse"
  }
}

`, connectionName)
}

func testAccDbtCloudSGlobalConnectionSynapseResourceFullConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  synapse = {
    host = "synapse.com"
    database = "synapse"

	// optional fields
	port = 1234
	retries = 3
	login_timeout = 1000
	query_timeout = 3600
  }
}
`, connectionName)
}

func TestAccDbtCloudGlobalConnectionStarburstResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionStarburstResourceBasicConfig(
					connectionName,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"trino_v0",
					),
				),
			},
			// modify, adding optional fields
			{
				Config: testAccDbtCloudSGlobalConnectionStarburstResourceFullConfig(
					connectionName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"trino_v0",
					),
				),
			},
			// IMPORT WITH ALL FIELDS
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// modify, removing optional fields to check PATCH when we remove fields
			{
				Config: testAccDbtCloudSGlobalConnectionStarburstResourceBasicConfig(
					connectionName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"trino_v0",
					),
				),
			},
			// IMPORT SUBSET
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})

}

func testAccDbtCloudSGlobalConnectionStarburstResourceBasicConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  starburst = {
    host = "starburst.com"
    database = "mydatabase"
  }
}

`, connectionName)
}

func testAccDbtCloudSGlobalConnectionStarburstResourceFullConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  starburst = {
    host = "starburst.com"
    database = "myotherdatabase"
  }
}
`, connectionName)
}

func TestAccDbtCloudGlobalConnectionAthenaResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionAthenaResourceBasicConfig(
					connectionName,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"athena_v0",
					),
				),
			},
			// modify, adding optional fields
			{
				Config: testAccDbtCloudSGlobalConnectionAthenaResourceFullConfig(
					connectionName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"athena_v0",
					),
				),
			},
			// IMPORT WITH ALL FIELDS
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// modify, removing optional fields to check PATCH when we remove fields
			{
				Config: testAccDbtCloudSGlobalConnectionAthenaResourceBasicConfig(
					connectionName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"athena_v0",
					),
				),
			},
			// IMPORT SUBSET
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})

}

func testAccDbtCloudSGlobalConnectionAthenaResourceBasicConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  athena = {
    region_name = "region"
    database = "database"
    s3_staging_dir = "s3://test-bucket/staging/"
  }
}

`, connectionName)
}

func testAccDbtCloudSGlobalConnectionAthenaResourceFullConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  athena = {
    region_name = "region"
    database = "database2"
    s3_staging_dir = "s3://test-bucket/other-staging/"
    work_group = "work_group" 
    spark_work_group = "spark_work_group"
    s3_data_dir = "s3://test-bucket/data/"
    s3_data_naming = "s3_data_naming"
    s3_tmp_table_dir = "s3://test-bucket/tmp/"
    poll_interval = 123
    num_retries = 2
    num_boto3_retries = 3
    num_iceberg_retries = 4 
  }
}
`, connectionName)
}

func TestAccDbtCloudGlobalConnectionApacheSparkResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionSparkResourceBasicConfig(
					connectionName,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"apache_spark_v0",
					),
				),
			},
			// modify, adding optional fields
			{
				Config: testAccDbtCloudSGlobalConnectionSparkResourceFullConfig(
					connectionName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"apache_spark_v0",
					),
				),
			},
			// IMPORT WITH ALL FIELDS
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// modify, removing optional fields to check PATCH when we remove fields
			{
				Config: testAccDbtCloudSGlobalConnectionSparkResourceBasicConfig(
					connectionName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"apache_spark_v0",
					),
				),
			},
			// IMPORT SUBSET
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})

}

func testAccDbtCloudSGlobalConnectionSparkResourceBasicConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`

resource dbtcloud_global_connection test {
  name = "%s"

  apache_spark = {
    method = "http"
    host = "spark.com"
    cluster = "cluster"
  }
}

`, connectionName)
}

func testAccDbtCloudSGlobalConnectionSparkResourceFullConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`
resource dbtcloud_global_connection test {
  name = "%s"

  apache_spark = {
    method = "thrift"
    host = "spark.com"
    cluster = "cluster"
	// optional fields
    port = 4321
    connect_timeout = 100
    connect_retries = 2
    organization = "org"
    user = "user"
    auth = "auth"
  }
}
`, connectionName)
}

func TestAccDbtCloudGlobalConnectionTeradataResource(t *testing.T) {

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionTeradataResourceBasicConfig(
					connectionName,
				),
				// we check the computed values, for the other ones the test suite already checks that the plan and state are the same
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"teradata_v0",
					),
				),
			},
			// modify, adding optional fields
			{
				Config: testAccDbtCloudSGlobalConnectionTeradataResourceBasicConfig(
					connectionName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"teradata_v0",
					),
				),
			},
			// IMPORT WITH ALL FIELDS
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			// modify, removing optional fields to check PATCH when we remove fields
			{
				Config: testAccDbtCloudSGlobalConnectionTeradataResourceBasicConfig(
					connectionName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"dbtcloud_global_connection.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_global_connection.test",
						"adapter_version",
						"teradata_v0",
					),
				),
			},
			// IMPORT SUBSET
			{
				ResourceName:            "dbtcloud_global_connection.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})

}

func testAccDbtCloudSGlobalConnectionTeradataResourceBasicConfig(
	connectionName string,
) string {
	return fmt.Sprintf(`

resource "dbtcloud_global_connection" test {
  name = "%s"

  teradata = {
    host       = "teradata.example.com"
	port       		= 1025
	tmode	   		= "ANSI"
	retries	   		= 3
	request_timeout = 3000
  }
}

`, connectionName)
}
