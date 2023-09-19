package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDbtCloudExtendedAttributesDataSource(t *testing.T) {

	config := extendedAttributes()

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.dbtcloud_extended_attributes.test", "extended_attributes_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_extended_attributes.test", "project_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_extended_attributes.test", "extended_attributes"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_environment.test", "extended_attributes_id"),
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

func extendedAttributes() string {
	return fmt.Sprintf(`
    resource "dbtcloud_project" "test_project" {
        name = "extended_attributes_test_project"
    }

    resource "dbtcloud_environment" "test_environment" {
        project_id = dbtcloud_project.test_project.id
        name = "extended_attributes_test_env"
        dbt_version = "%s"
        type = "development"
        extended_attributes_id = dbtcloud_extended_attributes.test.extended_attributes_id
    }

    resource "dbtcloud_extended_attributes" "test" {
        extended_attributes = jsonencode(
          {
            type      = "databricks"
            catalog   = "dbt_catalog"
            http_path = "/sql/your/http/path"
            my_nested_field = {
              subfield = "my_value"
            }
          }
        )
        project_id = dbtcloud_project.test_project.id
      }

    data "dbtcloud_extended_attributes" "test" {
        extended_attributes_id = dbtcloud_extended_attributes.test.extended_attributes_id
        project_id = dbtcloud_project.test_project.id
    }

    data "dbtcloud_environment" "test" {
      project_id = dbtcloud_project.test_project.id
      environment_id = dbtcloud_environment.test_environment.environment_id
  }
    `, DBT_CLOUD_VERSION)
}
