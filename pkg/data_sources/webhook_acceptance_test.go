package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDbtCloudWebhookDataSource(t *testing.T) {

	randomWebhookName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	randomWebhookDescription := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := webhooks(randomWebhookName, randomWebhookDescription)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.dbt_cloud_webhook.test", "webhook_id"),
		resource.TestCheckResourceAttr("data.dbt_cloud_webhook.test", "name", randomWebhookName),
		resource.TestCheckResourceAttr("data.dbt_cloud_webhook.test", "description", randomWebhookDescription),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_webhook.test", "client_url"),
		resource.TestCheckResourceAttr("data.dbt_cloud_webhook.test", "event_types.#", "2"),
		resource.TestCheckResourceAttr("data.dbt_cloud_webhook.test", "job_ids.#", "0"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_webhook.test", "http_status_code"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_webhook.test", "account_identifier"),
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
    resource "dbt_cloud_webhook" "test_webhook" {
        name = "%s"
        description = "%s"
        client_url = "http://localhost/nothing"
        event_types = [
            "job.run.started",
            "job.run.completed"
        ]
    }

    data "dbt_cloud_webhook" "test" {
        webhook_id = dbt_cloud_webhook.test_webhook.id
    }
    `, webhookName, webhookDesc)
}
