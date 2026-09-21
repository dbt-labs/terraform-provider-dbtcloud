// dbt Wizard, paid, with a spend limit of 500 USD.
// Paid activation charges the account for the usage of the product, and moves the
// account to a usage-based plan if it is not on one yet. Check the plan change in
// the output of `terraform plan` before you apply it.
resource "dbtcloud_account_add_on" "wizard" {
  product = "wizard"

  // 1 USD is 1,000,000,000 nanodollars. For dbt Wizard the limit stops usage once
  // the account reaches it.
  spend_limit_nanodollars = 500000000000
}

// dbt State, started as a trial. An account gets one trial for each product.
resource "dbtcloud_account_add_on" "state" {
  product    = "state"
  activation = "trial"
}

// dbt Wizard needs the AI features of the account, so turn them on first.
resource "dbtcloud_account_features" "my_features" {
  ai_features = true
}

resource "dbtcloud_account_add_on" "wizard_with_ai_features" {
  product = "wizard"

  depends_on = [dbtcloud_account_features.my_features]
}
