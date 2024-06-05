---
page_title: "dbtcloud_partial_notification Resource - dbtcloud"
subcategory: ""
description: |-
  Setup partial notifications on jobs success/failure to internal users, external email addresses or Slack channels. This is different from dbt_cloud_notification as it allows to have multiple resources updating the same notification recipient (email, user or Slack channel) and is useful for companies managing a single dbt Cloud Account configuration from different Terraform projects/workspaces.
  If a company uses only one Terraform project/workspace to manage all their dbt Cloud Account config, it is recommended to use dbt_cloud_notification instead of dbt_cloud_partial_notification.
  ~> This is a new resource. Feedback is welcome.
  The resource currently requires a Service Token with Account Admin access.
  The current behavior of the resource is the following:
  when using dbt_cloud_partial_notification, don't use dbt_cloud_notification for the same notification recipient in any other project/workspace. Otherwise, the behavior is undefined and partial notifications might be removed.when defining a new dbt_cloud_partial_notification
  if the notification recipient doesn't exist, it will be createdif a notification config exists for the current recipient, Job IDs will be added in the list of jobs to trigger the notificationsin a given Terraform project/workspace, avoid having different dbt_cloud_partial_notification for the same recipient to prevent sync issues. Add all the jobs in the same resource.all resources for the same notification recipient need to have the same values for state and user_id. Those fields are not considered "partial".when a resource is updated, the dbt Cloud notification recipient will be updated accordingly, removing and adding job ids in the list of jobs triggering notificationswhen the resource is deleted/destroyed, if the resulting notification recipient list of jobs is empty, the notification will be deleted ; otherwise, the notification will be updated, removing the job ids from the deleted resource
---

# dbtcloud_partial_notification (Resource)


Setup partial notifications on jobs success/failure to internal users, external email addresses or Slack channels. This is different from `dbt_cloud_notification` as it allows to have multiple resources updating the same notification recipient (email, user or Slack channel) and is useful for companies managing a single dbt Cloud Account configuration from different Terraform projects/workspaces.

If a company uses only one Terraform project/workspace to manage all their dbt Cloud Account config, it is recommended to use `dbt_cloud_notification` instead of `dbt_cloud_partial_notification`.

~> This is a new resource. Feedback is welcome.

The resource currently requires a Service Token with Account Admin access.

The current behavior of the resource is the following:

- when using `dbt_cloud_partial_notification`, don't use `dbt_cloud_notification` for the same notification recipient in any other project/workspace. Otherwise, the behavior is undefined and partial notifications might be removed.
- when defining a new `dbt_cloud_partial_notification`
  - if the notification recipient doesn't exist, it will be created
  - if a notification config exists for the current recipient, Job IDs will be added in the list of jobs to trigger the notifications
- in a given Terraform project/workspace, avoid having different `dbt_cloud_partial_notification` for the same recipient to prevent sync issues. Add all the jobs in the same resource. 
- all resources for the same notification recipient need to have the same values for `state` and `user_id`. Those fields are not considered "partial".
- when a resource is updated, the dbt Cloud notification recipient will be updated accordingly, removing and adding job ids in the list of jobs triggering notifications
- when the resource is deleted/destroyed, if the resulting notification recipient list of jobs is empty, the notification will be deleted ; otherwise, the notification will be updated, removing the job ids from the deleted resource

## Example Usage

```terraform
// the config is the same as for `dbtcloud_notification`

resource "dbtcloud_partial_notification" "prod_job_internal_notification" {
  // user_id is the internal ID of a given user in dbt Cloud
  user_id    = 100
  on_success = [dbtcloud_job.prod_job.id]
  on_failure = [12345]
  // the Type 1 is used for internal notifications
  notification_type = 1
}

// we can also send "external" email notifications to emails to related to dbt Cloud users
resource "dbtcloud_partial_notification" "prod_job_external_notification" {
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
resource "dbtcloud_partial_notification" "prod_job_slack_notifications" {
  // we still need the ID of a user in dbt Cloud even though it is not used for sending notifications
  user_id    = 100
  on_failure = [23456, 56788]
  on_cancel  = [dbtcloud_job.prod_job.id]
  // the Type 2 is used for Slack notifications
  notification_type  = 2
  slack_channel_id   = "C12345ABCDE"
  slack_channel_name = "#my-awesome-channel"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `user_id` (Number) Internal dbt Cloud User ID. Must be the user_id for an existing user even if the notification is an external one [global]

### Optional

- `external_email` (String) The external email to receive the notification [global, used as identifier]
- `notification_type` (Number) Type of notification (1 = dbt Cloud user email (default): does not require an external_email ; 2 = Slack channel: requires `slack_channel_id` and `slack_channel_name` ; 4 = external email: requires setting an `external_email`) [global, used as identifier]
- `on_cancel` (Set of Number) List of job IDs to trigger the webhook on cancel. Those will be added/removed when config is added/removed.
- `on_failure` (Set of Number) List of job IDs to trigger the webhook on failure Those will be added/removed when config is added/removed.
- `on_success` (Set of Number) List of job IDs to trigger the webhook on success Those will be added/removed when config is added/removed.
- `on_warning` (Set of Number) List of job IDs to trigger the webhook on warning Those will be added/removed when config is added/removed.
- `slack_channel_id` (String) The ID of the Slack channel to receive the notification. It can be found at the bottom of the Slack channel settings [global, used as identifier]
- `slack_channel_name` (String) The name of the slack channel [global, used as identifier]
- `state` (Number) State of the notification (1 = active (default), 2 = inactive) [global]

### Read-Only

- `id` (String) The ID of the notification
