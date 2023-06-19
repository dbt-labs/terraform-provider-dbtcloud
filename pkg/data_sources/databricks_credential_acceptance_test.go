package data_sources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudDatabricksCredentialDataSource(t *testing.T) {

	testDatabricks := os.Getenv("TEST_DATABRICKS")

	var adapterType string
	if testDatabricks == "true" {
		adapterType = "databricks"
	} else {
		adapterType = "spark"
	}
	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	config := databricks_credential(randomProjectName, "moo", "baa", "maa", 64, adapterType)

	// TODO: revisit when adapters can be created with a service token
	// as of now, CI is using a spark adapter and doesn't have a catalog
	// TEST_DATABRICKS is not set in CI
	var check resource.TestCheckFunc

	if testDatabricks == "true" {
		check = resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "credential_id"),
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "project_id"),
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "adapter_id"),
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "target_name"),
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "schema"),
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "num_threads"),
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "catalog"),
		)
	} else {
		check = resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "credential_id"),
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "project_id"),
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "adapter_id"),
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "target_name"),
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "schema"),
			resource.TestCheckResourceAttrSet("data.dbtcloud_databricks_credential.test", "num_threads"),
		)
	}

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

func databricks_credential(projectName string, defaultSchema string, username string, password string, numThreads int, adapterType string) string {
	commonConfig := fmt.Sprintf(`
    resource "dbtcloud_project" "test_credential_project" {
        name = "%s"
    }

    data "dbtcloud_databricks_credential" "test" {
        project_id = dbtcloud_project.test_credential_project.id
        credential_id = dbtcloud_databricks_credential.test_cred.credential_id
    }
    `, projectName)

	if adapterType == "databricks" {
		credential := `resource "dbtcloud_databricks_credential" "test_cred" {
			project_id = dbtcloud_project.test_credential_project.id
			adapter_id = 123
			token = "abcdefg"
			schema = "my_schema"
			adapter_type = "databricks"
			catalog = "my_catalog"
		}`

		return fmt.Sprintln(commonConfig, credential)
	} else {
		credential := `resource "dbtcloud_databricks_credential" "test_cred" {
			project_id = dbtcloud_project.test_credential_project.id
			adapter_id = 123
			token = "abcdefg"
			schema = "my_schema"
			adapter_type = "spark"
		}`
		return fmt.Sprintln(commonConfig, credential)
	}
}
