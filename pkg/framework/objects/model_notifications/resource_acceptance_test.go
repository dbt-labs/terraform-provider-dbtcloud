package model_notifications_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudModelNotificationsResource(t *testing.T) {
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudModelNotificationsResourceBasicConfig(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudModelNotificationsExists("dbtcloud_model_notifications.test_model_notifications"),
					resource.TestCheckResourceAttr(
						"dbtcloud_model_notifications.test_model_notifications",
						"enabled",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_model_notifications.test_model_notifications",
						"on_success",
						"false",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_model_notifications.test_model_notifications",
						"on_failure",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_model_notifications.test_model_notifications",
						"on_warning",
						"false",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_model_notifications.test_model_notifications",
						"on_skipped",
						"true",
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudModelNotificationsResourceModifyConfig(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudModelNotificationsExists("dbtcloud_model_notifications.test_model_notifications"),
					resource.TestCheckResourceAttr(
						"dbtcloud_model_notifications.test_model_notifications",
						"enabled",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_model_notifications.test_model_notifications",
						"on_success",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_model_notifications.test_model_notifications",
						"on_failure",
						"false",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_model_notifications.test_model_notifications",
						"on_warning",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_model_notifications.test_model_notifications",
						"on_skipped",
						"false",
					),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_model_notifications.test_model_notifications",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					// Get the environment ID from the state
					rs, ok := s.RootModule().Resources["dbtcloud_environment.test_environment"]
					if !ok {
						return "", fmt.Errorf("Not found: dbtcloud_environment.test_environment")
					}
					// Use environment_id attribute instead of ID
					return rs.Primary.Attributes["environment_id"], nil
				},
			},
		},
	})
}

func testAccDbtCloudModelNotificationsResourceBasicConfig(projectName string) string {
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
`, projectName)
}

func testAccDbtCloudModelNotificationsResourceModifyConfig(projectName string) string {
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
  on_success     = true
  on_failure     = false
  on_warning     = true
  on_skipped     = false
}
`, projectName)
}

func testAccCheckDbtCloudModelNotificationsExists(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		return nil
	}
}
