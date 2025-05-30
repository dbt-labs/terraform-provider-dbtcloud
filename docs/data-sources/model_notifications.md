---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dbtcloud_model_notifications Data Source - dbtcloud"
subcategory: ""
description: |-
  Get model notifications configuration for a dbt Cloud environment
---

# dbtcloud_model_notifications (Data Source)

Get model notifications configuration for a dbt Cloud environment

## Example Usage

```terraform
data "dbtcloud_model_notifications" "prod_model_notifications" {
  environment_id = dbtcloud_environment.prod_environment.environment_id
}

data "dbtcloud_model_notifications" "qa_model_notifications" {
  environment_id = 12345
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `environment_id` (String) The ID of the dbt Cloud environment

### Read-Only

- `enabled` (Boolean) Whether model notifications are enabled for this environment
- `id` (Number) The internal ID of the model notifications configuration
- `on_failure` (Boolean) Whether to send notifications for failed model runs
- `on_skipped` (Boolean) Whether to send notifications for skipped model runs
- `on_success` (Boolean) Whether to send notifications for successful model runs
- `on_warning` (Boolean) Whether to send notifications for model runs with warnings
