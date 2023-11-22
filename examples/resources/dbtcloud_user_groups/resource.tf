// we can assign groups to users
resource "dbtcloud_user_groups" "my_user_groups" {
  user_id = dbtcloud_user.my_user.id
  group_ids = [
    // the group_id can be written directly
    1234,
    // or we can refer to a group created by Terraform
    dbtcloud_group.my_group.id,
    // or we can use  a local variable (see the guide on how to use the HTTP provider)
    local.my_group_id,
  ]
}

// as Delete is not handled currently, by design, removing all groups from a user can be done with
resource "dbtcloud_user_groups" "my_other_user_groups" {
  user_id   = 123456
  group_ids = []
}