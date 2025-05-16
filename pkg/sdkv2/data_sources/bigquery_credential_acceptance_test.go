package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudBigQueryCredentialDataSource(t *testing.T) {

	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := bigquery_credential(randomProjectName, "moo", 64)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet(
			"data.dbtcloud_bigquery_credential.test",
			"credential_id",
		),
		resource.TestCheckResourceAttrSet("data.dbtcloud_bigquery_credential.test", "project_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_bigquery_credential.test", "dataset"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_bigquery_credential.test", "num_threads"),
	)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
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
    resource "dbtcloud_project" "test_credential_project" {
        name = "%s"
    }

    resource "dbtcloud_bigquery_credential" "test_cred" {
        project_id = dbtcloud_project.test_credential_project.id
        dataset = "%s"
        num_threads = "%d"
    }

    data "dbtcloud_bigquery_credential" "test" {
        project_id = dbtcloud_project.test_credential_project.id
        credential_id = dbtcloud_bigquery_credential.test_cred.credential_id
    }
    `, projectName, dataset, numThreads)
}
