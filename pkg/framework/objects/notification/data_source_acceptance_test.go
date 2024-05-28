package notification_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudNotificationDataSource(t *testing.T) {

	var userID string
	if acctest_helper.IsDbtCloudPR() {
		userID = "1"
	} else {
		userID = "100"
	}

	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := notification(randomProjectName, userID)

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
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  check,
			},
		},
	})
}

func notification(projectName, userID string) string {
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
		user_id           = %s
		on_failure        = [dbtcloud_job.test_notification_job_1.id]
		notification_type = 4
		external_email    = "nomail@mail.com"
	}

	data "dbtcloud_notification" "test_notification_external" {
		notification_id = dbtcloud_notification.test_notification_external.id
	}
    `, projectName, acctest_helper.DBT_CLOUD_VERSION, userID)
}
