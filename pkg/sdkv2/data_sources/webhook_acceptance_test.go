package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDbtCloudWebhookDataSource(t *testing.T) {

	randomWebhookName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	randomWebhookDescription := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := webhooks(randomWebhookName, randomWebhookDescription)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.dbtcloud_webhook.test", "webhook_id"),
		resource.TestCheckResourceAttr("data.dbtcloud_webhook.test", "name", randomWebhookName),
		resource.TestCheckResourceAttr(
			"data.dbtcloud_webhook.test",
			"description",
			randomWebhookDescription,
		),
		resource.TestCheckResourceAttrSet("data.dbtcloud_webhook.test", "client_url"),
		resource.TestCheckResourceAttr("data.dbtcloud_webhook.test", "event_types.#", "2"),
		resource.TestCheckResourceAttr("data.dbtcloud_webhook.test", "job_ids.#", "0"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_webhook.test", "http_status_code"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_webhook.test", "account_identifier"),
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

func webhooks(webhookName string, webhookDesc string) string {
	return fmt.Sprintf(`
    resource "dbtcloud_webhook" "test_webhook" {
        name = "%s"
        description = "%s"
        client_url = "http://localhost/nothing"
        event_types = [
            "job.run.started",
            "job.run.completed"
        ]
    }

    data "dbtcloud_webhook" "test" {
        webhook_id = dbtcloud_webhook.test_webhook.id
    }
    `, webhookName, webhookDesc)
}
