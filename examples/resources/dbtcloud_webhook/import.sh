# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_webhook.my_webhook
  id = "webhook_id"
}

import {
  to = dbtcloud_webhook.my_webhook
  id = "wsu_abcdefg"
}

# using the older import command
terraform import dbtcloud_webhook.my_webhook "webhook_id"
terraform import dbtcloud_webhook.my_webhook wsu_abcdefg
