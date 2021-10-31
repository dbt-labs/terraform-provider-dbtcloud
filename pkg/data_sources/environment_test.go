package data_sources_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudEnvironmentDataSource(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	randomIDInt, _ := strconv.Atoi(randomID)

	config := fmt.Sprintf(`
			data "dbt_cloud_environment" "test" {
				project_id = 123
				environment_id = %d
			}
		`, randomIDInt)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbt_cloud_environment.test", "environment_id", randomID),
		resource.TestCheckResourceAttr("data.dbt_cloud_environment.test", "project_id", "123"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "name"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "is_active"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "credential_id"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "dbt_version"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "type"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "use_custom_branch"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_project.test", "custom_branch"),
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
