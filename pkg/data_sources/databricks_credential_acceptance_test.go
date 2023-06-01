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

		// TODO: revisit when adapters can be created with a service token
		// as of now, CI is using a spark adapter and doesn't have a catalog
		// you can uncomment the following line to test locally on a databricks adapter

		// resource.TestCheckResourceAttrSet("data.dbt_cloud_databricks_credential.test", "catalog"),
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

// TODO: revisit when adapters can be created with a service token
// In CI, the Adapter 123 is of type "spark", but locally, for me it is databricks
// We can't create adapters right now with service tokens but should revisit when this is updated

func databricks_credential(projectName string, defaultSchema string, username string, password string, numThreads int) string {
	return fmt.Sprintf(`
    resource "dbt_cloud_project" "test_credential_project" {
        name = "%s"
    }

    resource "dbt_cloud_databricks_credential" "test_cred" {
        project_id = dbt_cloud_project.test_credential_project.id
		adapter_id = 123
        token = "abcdefg"
        schema = "my_schema"
		adapter_type = "spark"
		# adapter_type = "databricks"
        # catalog = "my_catalog"
    }

    data "dbt_cloud_databricks_credential" "test" {
        project_id = dbt_cloud_project.test_credential_project.id
        credential_id = dbt_cloud_databricks_credential.test_cred.credential_id
    }
    `, projectName)
}
