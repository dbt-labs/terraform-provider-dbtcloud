package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudSnowflakeCredentialDataSource(t *testing.T) {

	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := snowflake_credential(randomProjectName, "moo", "baa", "maa", 64)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "credential_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "project_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "auth_type"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "is_active"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "schema"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "user"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "num_threads"),
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

func snowflake_credential(projectName string, schema string, username string, password string, numThreads int) string {
	return fmt.Sprintf(`
    resource "dbt_cloud_project" "test_credential_project" {
        name = "%s"
    }

    resource "dbt_cloud_snowflake_credential" "test_cred" {
        project_id = dbt_cloud_project.test_credential_project.id
        num_threads = 64
        user = "moo"
        password = "baa"
        schema = "tst"
        auth_type = "password"
    }

    data "dbt_cloud_snowflake_credential" "test" {
        project_id = dbt_cloud_project.test_credential_project.id
        credential_id = dbt_cloud_snowflake_credential.test_cred.credential_id
    }
    `, projectName)
}
