package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudNotificationDataSource(t *testing.T) {

	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := notification(randomProjectName)

	check := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			"data.dbtcloud_notification.test_notification_external",
			"notification_type",
			"4",
		),
		resource.TestCheckResourceAttrSet(
			"data.dbtcloud_notification.test_notification_external",
			"on_failure.0",
		),
		resource.TestCheckResourceAttr(
			"data.dbtcloud_notification.test_notification_external",
			"external_email",
			"nomail@mail.com",
		),
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

func notification(projectName string) string {
	return fmt.Sprintf(`
	resource "dbtcloud_project" "test_notification_project" {
		name = "%s"
	}
		
	resource "dbtcloud_environment" "test_notification_environment" {
		project_id  = dbtcloud_project.test_notification_project.id
		name        = "Test Env Notification"
		dbt_version = "%s"
		type        = "development"
	}
		
	resource "dbtcloud_job" "test_notification_job_1" {
		name           = "Job 1 TF"
		project_id     = dbtcloud_project.test_notification_project.id
		environment_id = dbtcloud_environment.test_notification_environment.environment_id
		execute_steps = [
			"dbt compile"
		]
		triggers = {
			"github_webhook" : false,
			"git_provider_webhook" : false,
			"schedule" : false,
		}
	}

	resource "dbtcloud_notification" "test_notification_external" {
		user_id           = 100
		on_failure        = [dbtcloud_job.test_notification_job_1.id]
		notification_type = 4
		external_email    = "nomail@mail.com"
	}

	data "dbtcloud_notification" "test_notification_external" {
		notification_id = dbtcloud_notification.test_notification_external.id
	}
    `, projectName, DBT_CLOUD_VERSION)
}
