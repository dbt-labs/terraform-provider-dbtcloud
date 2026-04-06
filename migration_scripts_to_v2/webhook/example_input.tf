resource "dbtcloud_webhook" "job_completed" {
  client_url  = "https://example.com/webhook"
  event_types = ["job.run.completed"]
  name        = "Job Completed Hook"
  webhook_id  = "wsu_abc123"
}

resource "dbtcloud_webhook" "job_started" {
  client_url  = "https://example.com/webhook/start"
  event_types = ["job.run.started"]
  name        = "Job Started Hook"
  webhook_id  = "wsu_def456"
}

data "dbtcloud_webhook" "existing_hook" {
  id         = "wsu_abc123"
  webhook_id = "wsu_abc123"
}

output "webhook_endpoint" {
  value = data.dbtcloud_webhook.existing_hook.webhook_id
}
