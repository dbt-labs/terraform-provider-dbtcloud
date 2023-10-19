// use dbt_cloud_webhook instead of dbtcloud_webhook for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_webhook" "test_webhook" {
  name        = "test-webhook"
  description = "Test webhook"
  client_url  = "http://localhost/nothing"
  event_types = [
    "job.run.started",
    "job.run.completed"
  ]
  job_ids = [
    1234,
    5678
  ]
}
