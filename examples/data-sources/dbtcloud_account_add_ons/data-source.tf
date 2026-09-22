data "dbtcloud_account_add_ons" "all" {}

output "add_on_states" {
  value = { for add_on in data.dbtcloud_account_add_ons.all.add_ons : add_on.product => add_on.state }
}
