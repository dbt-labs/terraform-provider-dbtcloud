# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_notification.my_notification
  id = "notification_id"
}

import {
  to = dbtcloud_notification.my_notification
  id = "12345"
}

# using the older import command
terraform import dbtcloud_notification.my_notification "notification_id"
terraform import dbtcloud_notification.my_notification 12345
