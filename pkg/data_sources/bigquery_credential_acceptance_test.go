package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudBigQueryCredentialDataSource(t *testing.T) {

	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := bigquery_credential(randomProjectName, "moo", 64)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.dbt_cloud_bigquery_credential.test", "credential_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_bigquery_credential.test", "project_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_bigquery_credential.test", "dataset"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_bigquery_credential.test", "num_threads"),
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

func bigquery_credential(projectName string, dataset string, numThreads int) string {
	return fmt.Sprintf(`
    resource "dbt_cloud_project" "test_credential_project" {
        name = "%s"
    }

    resource "dbt_cloud_bigquery_credential" "test_cred" {
        project_id = dbt_cloud_project.test_credential_project.id
        dataset = "%s"
        num_threads = "%d"
    }

    data "dbt_cloud_bigquery_credential" "test" {
        project_id = dbt_cloud_project.test_credential_project.id
        credential_id = dbt_cloud_bigquery_credential.test_cred.credential_id
    }
    `, projectName, dataset, numThreads)
}
