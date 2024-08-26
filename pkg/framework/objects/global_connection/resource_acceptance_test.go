package global_connection_test

import (
	"fmt"
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

	resource.Test(t, resource.TestCase{
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
resource dbtcloud_global_connection test {
  name = "%s"

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

	if acctest_helper.IsDbtCloudPR() {
		// TODO: remove this when global connections is on everywhere
		t.Skip("Skipping global connections in dbt Cloud CI for now")
	}

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with just mandatory fields
			{
				Config: testAccDbtCloudSGlobalConnectionBigQueryResourceBasicConfig(
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
				),
			},
			// modify, removing optional fields to check PATCH when we remove fields
			{
				Config: testAccDbtCloudSGlobalConnectionBigQueryResourceBasicConfig(
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

func testAccDbtCloudSGlobalConnectionBigQueryResourceBasicConfig(
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

  }
}

`, connectionName)
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

  }
}
`, connectionName)
}

func TestAccDbtCloudGlobalConnectionDatabricksResource(t *testing.T) {

	if acctest_helper.IsDbtCloudPR() {
		// TODO: remove this when global connections is on everywhere
		t.Skip("Skipping global connections in dbt Cloud CI for now")
	}

	connectionName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	connectionName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	oAuthClientID := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	oAuthClientSecret := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
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
