data "dbtcloud_account_add_on" "wizard" {
  product = "wizard"
}

output "wizard_state" {
  value = data.dbtcloud_account_add_on.wizard.state
}
