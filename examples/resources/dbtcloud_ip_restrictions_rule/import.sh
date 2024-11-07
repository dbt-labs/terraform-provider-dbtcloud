# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_ip_restrictions_rule.my_rule
  id = "ip_restriction_rule_id"
}

import {
  to = dbtcloud_ip_restrictions_rule.my_rule
  id = "12345"
}

# using the older import command
terraform import dbtcloud_ip_restrictions_rule.my_rule "ip_restriction_rule_id"
terraform import dbtcloud_ip_restrictions_rule.my_rule 12345
