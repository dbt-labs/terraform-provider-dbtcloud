package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudBigQueryConnectionDataSource(t *testing.T) {

	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	randomConnectionName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := bigQueryConnection(randomProjectName, randomConnectionName)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "name", randomConnectionName),
		resource.TestCheckResourceAttrSet("data.dbtcloud_bigquery_connection.test", "connection_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_bigquery_connection.test", "project_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_bigquery_connection.test", "is_active"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_bigquery_connection.test", "type"),

		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "gcp_project_id", "test_gcp_project_id"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "timeout_seconds", "100"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "private_key_id", "test_private_key_id"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "client_email", "test_client_email"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "client_id", "test_client_id"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "auth_uri", "test_auth_uri"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "token_uri", "test_token_uri"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "auth_provider_x509_cert_url", "test_auth_provider_x509_cert_url"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "client_x509_cert_url", "test_client_x509_cert_url"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "retries", "3"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "location", "EU"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "maximum_bytes_billed", "100000"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "gcs_bucket", "test_gcs_bucket"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "dataproc_region", "test_dataproc_region"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "dataproc_cluster_name", "test_dataproc_cluster_name"),
		resource.TestCheckResourceAttr("data.dbtcloud_bigquery_connection.test", "is_configured_for_oauth", "false"),
	)

	resource.ParallelTest(t, resource.TestCase{
		Providers: providers(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  check,
			},
		},
	})
}

func bigQueryConnection(projectName, connectionName string) string {
	return fmt.Sprintf(`
    resource "dbtcloud_project" "test_project" {
        name = "%s"
    }

    resource "dbtcloud_bigquery_connection" "test_connection" {
        project_id = dbtcloud_project.test_project.id
        name = "%s"
        type = "bigquery"
        is_active = true
		gcp_project_id = "test_gcp_project_id"
		timeout_seconds = 100
		private_key_id = "test_private_key_id"
		private_key = "XXX"
		client_email = "test_client_email"
		client_id = "test_client_id"
		auth_uri = "test_auth_uri"
		token_uri = "test_token_uri"
		auth_provider_x509_cert_url = "test_auth_provider_x509_cert_url"
		client_x509_cert_url = "test_client_x509_cert_url"
		retries = 3
		location = "EU"
		maximum_bytes_billed = 100000
		gcs_bucket = "test_gcs_bucket"
		dataproc_region = "test_dataproc_region"
		dataproc_cluster_name = "test_dataproc_cluster_name"
    }

    data "dbtcloud_bigquery_connection" "test" {
        project_id = dbtcloud_project.test_project.id
        connection_id = dbtcloud_bigquery_connection.test_connection.connection_id
    }
    `, projectName, connectionName)
}
