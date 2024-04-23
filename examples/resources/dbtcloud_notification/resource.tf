// dbt Cloud allows us to create internal and external notifications
//
// an internal notification will send emails to the user mentioned in `user_id`
//
// NOTE: If internal notification settings already exist for a user, currently you MUST import
// those first into the state file before you can create a new internal notification for that user.
// Failure to do so, will result in the user losing access to existing notifications and dbt
// support will need to be contacted to restore access.
// cmd: terraform import dbtcloud_notification.prod_job_internal_notification <user_id>

resource "dbtcloud_notification" "prod_job_internal_notification" {
  // user_id is the internal ID of a given user in dbt Cloud
  user_id    = 100
  on_success = [dbtcloud_job.prod_job.id]
  on_failure = [12345]
  // the Type 1 is used for internal notifications
  notification_type = 1
}

// we can also send "external" email notifications to emails to related to dbt Cloud users
resource "dbtcloud_notification" "prod_job_external_notification" {
  // we still need the ID of a user in dbt Cloud even though it is not used for sending notifications
  user_id    = 100
  on_failure = [23456, 56788]
  on_cancel  = [dbtcloud_job.prod_job.id]
  // the Type 4 is used for external notifications
  notification_type = 4
  // the external_email is the email address that will receive the notification
  external_email = "my_email@mail.com"
}

// and finally, we can set up Slack notifications
resource "dbtcloud_notification" "prod_job_slack_notifications" {
  // we still need the ID of a user in dbt Cloud even though it is not used for sending notifications
  user_id    = 100
  on_failure = [23456, 56788]
  on_cancel  = [dbtcloud_job.prod_job.id]
  // the Type 2 is used for Slack notifications
  notification_type  = 2
  slack_channel_id   = "C12345ABCDE"
  slack_channel_name = "#my-awesome-channel"
}