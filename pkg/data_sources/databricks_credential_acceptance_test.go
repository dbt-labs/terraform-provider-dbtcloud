package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudDatabricksCredentialDataSource(t *testing.T) {

	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := databricks_credential(randomProjectName, "moo", "baa", "maa", 64)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.dbt_cloud_databricks_credential.test", "credential_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_databricks_credential.test", "project_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_databricks_credential.test", "adapter_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_databricks_credential.test", "target_name"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_databricks_credential.test", "schema"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_databricks_credential.test", "num_threads"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_databricks_credential.test", "catalog"),
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

func databricks_credential(projectName string, defaultSchema string, username string, password string, numThreads int) string {
	return fmt.Sprintf(`
    resource "dbt_cloud_project" "test_credential_project" {
        name = "%s"
    }

    resource "dbt_cloud_databricks_credential" "test_cred" {
        project_id = dbt_cloud_project.test_credential_project.id
        num_threads = 32
		adapter_id = 123
        token = "abcdefg"
        catalog = "my_catalog"
        schema = "my_schema"
    }

    data "dbt_cloud_databricks_credential" "test" {
        project_id = dbt_cloud_project.test_credential_project.id
        credential_id = dbt_cloud_databricks_credential.test_cred.credential_id
    }
    `, projectName)
}
