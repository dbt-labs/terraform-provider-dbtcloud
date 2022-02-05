package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudConnectionDataSource(t *testing.T) {

	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	randomConnectionName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := connection(randomProjectName, randomConnectionName)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbt_cloud_connection.test", "name", randomConnectionName),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_connection.test", "connection_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_connection.test", "project_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_connection.test", "is_active"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_connection.test", "type"),
		resource.TestCheckResourceAttr("data.dbt_cloud_connection.test", "account", "test_account"),
		resource.TestCheckResourceAttr("data.dbt_cloud_connection.test", "database", "test_db"),
		resource.TestCheckResourceAttr("data.dbt_cloud_connection.test", "warehouse", "test_wh"),
		resource.TestCheckResourceAttr("data.dbt_cloud_connection.test", "role", "test_role"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_connection.test", "allow_sso"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_connection.test", "allow_keep_alive"),
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

func connection(projectName, connectionName string) string {
	return fmt.Sprintf(`
    resource "dbt_cloud_project" "test_project" {
        name = "%s"
    }

    resource "dbt_cloud_connection" "test_connection" {
        project_id = dbt_cloud_project.test_project.id
        name = "%s"
        type = "snowflake"
        is_active = true
        account = "test_account"
        database = "test_db"
        warehouse = "test_wh"
        role = "test_role"
        allow_sso = true
        allow_keep_alive = true
    }

    data "dbt_cloud_connection" "test" {
        project_id = dbt_cloud_project.test_project.id
        connection_id = dbt_cloud_connection.test_connection.connection_id
    }
    `, projectName, connectionName)
}
