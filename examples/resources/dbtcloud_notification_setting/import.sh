# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_notification_setting.my_setting
  id = "notification_setting_id"
}

import {
  to = dbtcloud_notification_setting.my_setting
  id = "12345"
}

# using the older import command
terraform import dbtcloud_notification_setting.my_setting "notification_setting_id"
terraform import dbtcloud_notification_setting.my_setting 12345
