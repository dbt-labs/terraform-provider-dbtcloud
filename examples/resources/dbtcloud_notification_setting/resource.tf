// dbtcloud_notification_setting configures Microsoft Teams and webhook notifications
// using dbt Cloud's notifications system.
//
// It is a separate resource from `dbtcloud_notification` because Teams is delivered
// through a redesigned notifications system that models a single setting as a
// collection of channels (where to send) and rules (when to send). Email and Slack
// notifications still use the legacy `dbtcloud_notification` resource.

// Microsoft Teams: notify on errors for a specific job
resource "dbtcloud_notification_setting" "teams_prod_failures" {
  name        = "Prod failures to Teams"
  description = "Alerts the data platform Teams channel when production runs error."

  channels = [
    {
      channel_type     = "teams"
      teams_team_id    = "19:abcdef0123456789@thread.tacv2"
      teams_channel_id = "19:fedcba9876543210@thread.tacv2"
    },
  ]

  rules = [
    {
      trigger_on = "run_errored"
      job_id     = dbtcloud_job.prod_job.id
    },
    {
      trigger_on = "run_cancelled"
      job_id     = dbtcloud_job.prod_job.id
    },
  ]
}

// Webhook: fan all-job warnings out to a custom endpoint
resource "dbtcloud_notification_setting" "webhook_warnings" {
  name      = "All-job warnings to PagerDuty"
  is_active = true

  channels = [
    {
      channel_type        = "webhook"
      webhook_client_url  = "https://events.pagerduty.com/integration/abc123/enqueue"
      webhook_hmac_secret = var.pagerduty_hmac_secret
    },
  ]

  rules = [
    // Omit job_id to fire for all jobs in the account.
    { trigger_on = "run_warning" },
  ]
}
