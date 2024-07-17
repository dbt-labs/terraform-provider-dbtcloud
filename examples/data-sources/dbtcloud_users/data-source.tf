// return all users in the dbt Cloud account
data "dbtcloud_users" "all" {
}

// we can use it to check if a user exists or not
// the dbtcloud_user datasource would fail if the user doesn't exist 
locals {
  user_details = [for user in data.dbtcloud_users.all.users : user if user.email == "example@amail.com"]
  user_exist   = length(local.user_details) == 1
}
