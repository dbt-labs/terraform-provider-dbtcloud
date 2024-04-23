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
