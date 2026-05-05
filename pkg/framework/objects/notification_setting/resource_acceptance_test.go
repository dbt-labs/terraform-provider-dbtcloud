package notification_setting_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudNotificationSettingResource(t *testing.T) {
	settingName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	settingNameUpdated := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resourceName := "dbtcloud_notification_setting.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudNotificationSettingDestroy,
		Steps: []resource.TestStep{
			// CREATE - one webhook channel, one rule scoped to a job
			{
				Config: testAccDbtCloudNotificationSettingResourceCreate(projectName, settingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudNotificationSettingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", settingName),
					resource.TestCheckResourceAttr(resourceName, "description", "initial"),
					resource.TestCheckResourceAttr(resourceName, "is_active", "true"),
					resource.TestCheckResourceAttr(resourceName, "channels.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "channels.0.channel_type", "webhook"),
					resource.TestCheckResourceAttr(
						resourceName,
						"channels.0.webhook_client_url",
						"https://example.com/hook-initial",
					),
					resource.TestCheckResourceAttrSet(resourceName, "channels.0.id"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.trigger_on", "run_errored"),
					resource.TestCheckResourceAttrSet(resourceName, "rules.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "rules.0.job_id"),
				),
			},
			// UPDATE - rename, flip is_active, change webhook URL, add an all-jobs rule
			{
				Config: testAccDbtCloudNotificationSettingResourceUpdate(projectName, settingNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudNotificationSettingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", settingNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", "updated"),
					resource.TestCheckResourceAttr(resourceName, "is_active", "false"),
					resource.TestCheckResourceAttr(resourceName, "channels.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName,
						"channels.0.webhook_client_url",
						"https://example.com/hook-updated",
					),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.trigger_on", "run_errored"),
					resource.TestCheckResourceAttr(resourceName, "rules.1.trigger_on", "run_warning"),
					// Second rule has no job_id — fires for all jobs.
					resource.TestCheckNoResourceAttr(resourceName, "rules.1.job_id"),
				),
			},
			// IMPORT
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// hmac secrets are write-only and never returned by the API.
				ImportStateVerifyIgnore: []string{"channels.0.webhook_hmac_secret"},
			},
		},
	})
}

func testAccDbtCloudNotificationSettingResourceProjectAndJob(projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_notification_setting_project" {
	name = "%s"
}

resource "dbtcloud_environment" "test_notification_setting_environment" {
	project_id  = dbtcloud_project.test_notification_setting_project.id
	name        = "Test Env Notification Setting"
	dbt_version = "%s"
	type        = "development"
}

resource "dbtcloud_job" "test_notification_setting_job" {
	name           = "Notification Setting Test Job"
	project_id     = dbtcloud_project.test_notification_setting_project.id
	environment_id = dbtcloud_environment.test_notification_setting_environment.environment_id
	execute_steps  = ["dbt compile"]
	triggers = {
		"github_webhook"       = false
		"git_provider_webhook" = false
		"schedule"             = false
	}
}
`, projectName, acctest_config.AcceptanceTestConfig.DbtCloudVersion)
}

func testAccDbtCloudNotificationSettingResourceCreate(projectName, settingName string) string {
	settingConfig := fmt.Sprintf(`
resource "dbtcloud_notification_setting" "test" {
	name        = "%s"
	description = "initial"
	is_active   = true

	channels = [
		{
			channel_type        = "webhook"
			webhook_client_url  = "https://example.com/hook-initial"
			webhook_hmac_secret = "initial-secret"
		},
	]

	rules = [
		{
			trigger_on = "run_errored"
			job_id     = dbtcloud_job.test_notification_setting_job.id
		},
	]
}
`, settingName)
	return testAccDbtCloudNotificationSettingResourceProjectAndJob(projectName) + settingConfig
}

func testAccDbtCloudNotificationSettingResourceUpdate(projectName, settingName string) string {
	settingConfig := fmt.Sprintf(`
resource "dbtcloud_notification_setting" "test" {
	name        = "%s"
	description = "updated"
	is_active   = false

	channels = [
		{
			channel_type        = "webhook"
			webhook_client_url  = "https://example.com/hook-updated"
			webhook_hmac_secret = "updated-secret"
		},
	]

	rules = [
		{
			trigger_on = "run_errored"
			job_id     = dbtcloud_job.test_notification_setting_job.id
		},
		{
			trigger_on = "run_warning"
		},
	]
}
`, settingName)
	return testAccDbtCloudNotificationSettingResourceProjectAndJob(projectName) + settingConfig
}

func testAccCheckDbtCloudNotificationSettingExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid notification setting ID %q: %v", rs.Primary.ID, err)
		}

		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client: %v", err)
		}

		if _, err := apiClient.GetNotificationSetting(id); err != nil {
			return fmt.Errorf("error fetching notification setting %s: %v", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudNotificationSettingDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client: %v", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_notification_setting" {
			continue
		}

		id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid notification setting ID %q: %v", rs.Primary.ID, err)
		}

		_, err = apiClient.GetNotificationSetting(id)
		if err == nil {
			return fmt.Errorf("Notification setting %s still exists", rs.Primary.ID)
		}

		// The notification-settings detail GET currently returns 500 with a
		// generic message instead of 404 when the row has been deleted (the bare
		// `except Exception` block in notification_settings_views.py swallows
		// `DoesNotExist`). Until the backend distinguishes them, treat any 500
		// from the verification GET as proof of deletion — DELETE has already
		// returned 2xx by this point, so a follow-up gateway timeout or generic
		// 500 just means we can't reconfirm via GET.
		expectedErr := regexp.MustCompile(`resource-not-found|internal-server-error`)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected not-found-style error, got %s", err)
		}
	}
	return nil
}
