package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudEnvironmentResource(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := fmt.Sprintf(`
			resource "dbt_cloud_environment" "test" {
				is_active = true
				name = "dbt-cloud-environment-%s"
				project_id = 123
				dbt_version = "0.21.0"
				type = "deployment"
				use_custom_branch = true
				custom_branch = "dev"
			}
		`, randomID)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("dbt_cloud_environment.test", "environment_id"),
		resource.TestCheckResourceAttr("dbt_cloud_environment.test", "is_active", "true"),
		resource.TestCheckResourceAttr("dbt_cloud_environment.test", "name", fmt.Sprintf("dbt-cloud-job-%s", randomID)),
		resource.TestCheckResourceAttr("dbt_cloud_environment.test", "project_id", "123"),
		resource.TestCheckResourceAttr("dbt_cloud_environment.test", "dbt_version", "0.21.0"),
		resource.TestCheckResourceAttr("dbt_cloud_environment.test", "type", "deployment"),
		resource.TestCheckResourceAttr("dbt_cloud_environment.test", "use_custom_branch", "true"),
		resource.TestCheckResourceAttr("dbt_cloud_environment.test", "custom_branch", "dev"),
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
