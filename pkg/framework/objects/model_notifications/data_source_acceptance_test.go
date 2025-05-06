package model_notifications_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudModelNotificationsDataSource(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudModelNotificationsDataSourceConfig(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.dbtcloud_model_notifications.test_model_notifications_ds",
						"id",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_model_notifications.test_model_notifications_ds",
						"enabled",
						"true",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_model_notifications.test_model_notifications_ds",
						"on_success",
						"false",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_model_notifications.test_model_notifications_ds",
						"on_failure",
						"true",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_model_notifications.test_model_notifications_ds",
						"on_warning",
						"false",
					),
					resource.TestCheckResourceAttr(
						"data.dbtcloud_model_notifications.test_model_notifications_ds",
						"on_skipped",
						"true",
					),
				),
			},
		},
	})
}

func testAccDbtCloudModelNotificationsDataSourceConfig(projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_project" {
  name = "%s"
}

resource "dbtcloud_environment" "test_environment" {
  project_id = dbtcloud_project.test_project.id
  name       = "Test Environment"
  dbt_version = "latest"
  type = "deployment"
}

resource "dbtcloud_model_notifications" "test_model_notifications" {
  environment_id = dbtcloud_environment.test_environment.environment_id
  enabled        = true
  on_success     = false
  on_failure     = true
  on_warning     = false
  on_skipped     = true
}

data "dbtcloud_model_notifications" "test_model_notifications_ds" {
  environment_id = dbtcloud_environment.test_environment.environment_id
  depends_on = [dbtcloud_model_notifications.test_model_notifications]
}
`, projectName)
}
