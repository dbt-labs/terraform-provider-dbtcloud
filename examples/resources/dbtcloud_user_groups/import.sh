# Import using the User ID
# The User ID can be retrieved from the dbt Cloud UI or with the data source dbtcloud_user
terraform import dbtcloud_user_groups.my_user_groups "user_id"
terraform import dbtcloud_user_groups.my_user_groups 123456
