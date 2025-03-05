# Model notifications are imported using the environment ID where the notifications are enabled
# Using import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_model_notifications.my_model_notifications
  id = "environment_id"
}

import {
  to = dbtcloud_model_notifications.my_model_notifications
  id = "12345"
}

# Using the older import command
terraform import dbtcloud_model_notifications.my_model_notifications "environment_id"
terraform import dbtcloud_model_notifications.my_model_notifications 12345 