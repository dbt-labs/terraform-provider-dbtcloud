resource "dbtcloud_ip_restrictions_rule" "test" {
  name        = "My restriction rule"
  description = "Important description"
  cidrs = [
    {
      cidr = "::ffff:106:708" # IPv6 config
    },
    {
      cidr = "1.6.7.10/24" # /24 for adding a range of addresses via netmask
    }
  ]
  type             = "deny"
  rule_set_enabled = false
}