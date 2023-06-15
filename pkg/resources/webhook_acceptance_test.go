package resources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudWebhookResource(t *testing.T) {

	webhookName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	webhookName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudWebhookResourceBasicConfig(webhookName, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudWebhookExists("dbt_cloud_webhook.test_webhook"),
					resource.TestCheckResourceAttr("dbt_cloud_webhook.test_webhook", "name", webhookName),
					resource.TestCheckResourceAttrSet("dbt_cloud_webhook.test_webhook", "hmac_secret"),
					resource.TestCheckResourceAttrSet("dbt_cloud_webhook.test_webhook", "account_identifier"),
					resource.TestCheckResourceAttr("dbt_cloud_webhook.test_webhook", "event_types.#", "2"),
					resource.TestCheckResourceAttr("dbt_cloud_webhook.test_webhook", "job_ids.#", "0"),
					resource.TestCheckResourceAttr("dbt_cloud_webhook.test_webhook", "client_url", "http://localhost/nothing"),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudWebhookResourceFullConfig(webhookName2, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudWebhookExists("dbt_cloud_webhook.test_webhook"),
					resource.TestCheckResourceAttr("dbt_cloud_webhook.test_webhook", "name", webhookName2),
					resource.TestCheckResourceAttrSet("dbt_cloud_webhook.test_webhook", "hmac_secret"),
					resource.TestCheckResourceAttrSet("dbt_cloud_webhook.test_webhook", "account_identifier"),
					resource.TestCheckResourceAttr("dbt_cloud_webhook.test_webhook", "event_types.#", "1"),
					resource.TestCheckResourceAttr("dbt_cloud_webhook.test_webhook", "job_ids.#", "1"),
					resource.TestCheckResourceAttr("dbt_cloud_webhook.test_webhook", "client_url", "http://localhost/new-nothing"),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbt_cloud_webhook.test_webhook",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"hmac_secret",
				},
			},
		},
	})
}

func testAccDbtCloudWebhookResourceBasicConfig(webhookName, projectName string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "%s"
}
resource "dbt_cloud_webhook" "test_webhook" {
	name = "%s"
	description = "My webhook"
	client_url = "http://localhost/nothing"
	event_types = [
	  "job.run.started",
	  "job.run.completed"
	]
  }
`, projectName, webhookName)
}

func testAccDbtCloudWebhookResourceFullConfig(webhookName, projectName string) string {
	return fmt.Sprintf(`
resource "dbt_cloud_project" "test_project" {
  name        = "%s"
}
resource "dbt_cloud_environment" "test_environment" {
	dbt_version   = "1.0.1"
	name          = "test"
	project_id    = dbt_cloud_project.test_project.id
	type          = "deployment"
  }
resource "dbt_cloud_job" "test" {
	environment_id = dbt_cloud_environment.test_environment.environment_id
	execute_steps = [
	  "dbt test"
	]
	generate_docs        = false
	is_active            = true
	name                 = "Test"
	num_threads          = 64
	project_id           = dbt_cloud_project.test_project.id
	run_generate_sources = false
	target_name          = "default"
	triggers = {
	  "custom_branch_only" : false,
	  "github_webhook" : false,
	  "git_provider_webhook" : false,
	  "schedule" : false
	}
  }
resource "dbt_cloud_webhook" "test_webhook" {
	name = "%s"
	description = "My webhook"
	client_url = "http://localhost/new-nothing"
	event_types = [
	  "job.run.completed"
	]
	job_ids = [dbt_cloud_job.test.id]
  }
`, projectName, webhookName)
}

func testAccCheckDbtCloudWebhookExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		webhookID := rs.Primary.ID

		_, err := apiClient.GetWebhook(webhookID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudWebhookDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbt_cloud_webhook" {
			continue
		}
		webhookID := rs.Primary.ID
		_, err := apiClient.GetWebhook(webhookID)
		if err == nil {
			return fmt.Errorf("Webhook still exists")
		}
		notFoundErr := "not found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
