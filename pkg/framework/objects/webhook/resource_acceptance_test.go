package webhook_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var webhookName = acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
var webhookName2 = acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
var projectName = acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
var active = "false"

var basicConfigTestStep = resource.TestStep{
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
			"https://example.com",
		),
	),
}

var modifyConfigTestStep = resource.TestStep{
	Config: testAccDbtCloudWebhookResourceFullConfig(webhookName2, projectName, "https://example.com/test", active),
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
			"https://example.com/test",
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_webhook.test_webhook",
			"active",
			active,
		),
	),
}

// resurrectionPlanErrorTestStep verifies that changing client_url on an
// inactive webhook while keeping active=false fails at plan time with a
// clear error, because the dbt Cloud API would reactivate the webhook and
// produce an inconsistent result.
var resurrectionPlanErrorTestStep = resource.TestStep{
	Config:      testAccDbtCloudWebhookResourceFullConfig(webhookName2, projectName, "https://example.com/resurrected", "false"),
	ExpectError: regexp.MustCompile(`Cannot change client_url while active=false`),
}

// resurrectionConvergeTestStep is the happy path: the user opts into
// reactivation by setting active=true alongside the client_url change.
var resurrectionConvergeTestStep = resource.TestStep{
	Config: testAccDbtCloudWebhookResourceFullConfig(webhookName2, projectName, "https://example.com/resurrected", "true"),
	Check: resource.ComposeTestCheckFunc(
		testAccCheckDbtCloudWebhookExists("dbtcloud_webhook.test_webhook"),
		resource.TestCheckResourceAttr(
			"dbtcloud_webhook.test_webhook",
			"client_url",
			"https://example.com/resurrected",
		),
		resource.TestCheckResourceAttr(
			"dbtcloud_webhook.test_webhook",
			"active",
			"true",
		),
	),
}

func TestAccDbtCloudWebhookResource(t *testing.T) {
	importStateTestStep := resource.TestStep{
		ResourceName:      "dbtcloud_webhook.test_webhook",
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"hmac_secret",
		},
	}

	// test the Framework implementation
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudWebhookDestroy,
		Steps: []resource.TestStep{
			basicConfigTestStep,
			modifyConfigTestStep,
			resurrectionPlanErrorTestStep,
			resurrectionConvergeTestStep,
			importStateTestStep,
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
	client_url = "https://example.com"
	event_types = [
	  "job.run.started",
	  "job.run.completed"
	]
  }
`, projectName, webhookName)
}

func testAccDbtCloudWebhookResourceFullConfig(webhookName, projectName, clientURL, active string) string {
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
	client_url = "%s"
	event_types = [
	  "job.run.completed"
	]
	job_ids = [dbtcloud_job.test.id]
	active = "%s"
  }
`, projectName, acctest_config.DBT_CLOUD_VERSION, webhookName, clientURL, active)
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
		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		webhookID := rs.Primary.ID

		_, err = apiClient.GetWebhook(webhookID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudWebhookDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

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
