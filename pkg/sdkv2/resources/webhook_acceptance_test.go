package resources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					testAccCheckDbtCloudWebhookExists("dbtcloud_webhook.test_webhook"),
					resource.TestCheckResourceAttr(
						"dbtcloud_webhook.test_webhook",
						"name",
						webhookName,
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_webhook.test_webhook",
						"hmac_secret",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_webhook.test_webhook",
						"account_identifier",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_webhook.test_webhook",
						"event_types.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_webhook.test_webhook",
						"job_ids.#",
						"0",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_webhook.test_webhook",
						"client_url",
						"http://localhost/nothing",
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudWebhookResourceFullConfig(webhookName2, projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudWebhookExists("dbtcloud_webhook.test_webhook"),
					resource.TestCheckResourceAttr(
						"dbtcloud_webhook.test_webhook",
						"name",
						webhookName2,
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_webhook.test_webhook",
						"hmac_secret",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_webhook.test_webhook",
						"account_identifier",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_webhook.test_webhook",
						"event_types.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_webhook.test_webhook",
						"job_ids.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_webhook.test_webhook",
						"client_url",
						"http://localhost/new-nothing",
					),
				),
			},
			// IMPORT
			{
				ResourceName:      "dbtcloud_webhook.test_webhook",
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
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_webhook" "test_webhook" {
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
resource "dbtcloud_project" "test_project" {
  name        = "%s"
}
resource "dbtcloud_environment" "test_environment" {
	dbt_version   = "%s"
	name          = "test"
	project_id    = dbtcloud_project.test_project.id
	type          = "deployment"
  }
resource "dbtcloud_job" "test" {
	environment_id = dbtcloud_environment.test_environment.environment_id
	execute_steps = [
	  "dbt test"
	]
	generate_docs        = false
	is_active            = true
	name                 = "Test"
	num_threads          = 64
	project_id           = dbtcloud_project.test_project.id
	run_generate_sources = false
	target_name          = "default"
	triggers = {
	  "github_webhook" : false,
	  "git_provider_webhook" : false,
	  "schedule" : false
	}
  }
resource "dbtcloud_webhook" "test_webhook" {
	name = "%s"
	description = "My webhook"
	client_url = "http://localhost/new-nothing"
	event_types = [
	  "job.run.completed"
	]
	job_ids = [dbtcloud_job.test.id]
  }
`, projectName, DBT_CLOUD_VERSION, webhookName)
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
		if rs.Type != "dbtcloud_webhook" {
			continue
		}
		webhookID := rs.Primary.ID
		_, err := apiClient.GetWebhook(webhookID)
		if err == nil {
			return fmt.Errorf("Webhook still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
