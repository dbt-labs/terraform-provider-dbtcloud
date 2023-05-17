resource "dbt_cloud_webhook" "test_webhook" {
  name = "my-webhook"
  description = "My webhook"
  client_url = "http://localhost/nothing"
  event_types = [
    "job.run.started",
    "job.run.completed"
  ]
  job_ids = [
    1234,
    5678
  ]
}