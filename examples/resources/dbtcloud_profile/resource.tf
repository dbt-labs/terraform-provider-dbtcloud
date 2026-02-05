# A profile ties together a connection and credentials for use within environments.
resource "dbtcloud_profile" "my_profile" {
  project_id     = dbtcloud_project.my_project.id
  key            = "my-profile"
  connection_id  = dbtcloud_global_connection.my_connection.id
  credentials_id = dbtcloud_snowflake_credential.my_credential.credential_id
}

# A profile with extended attributes
resource "dbtcloud_profile" "my_profile_with_attrs" {
  project_id              = dbtcloud_project.my_project.id
  key                     = "my-profile-with-attrs"
  connection_id           = dbtcloud_global_connection.my_connection.id
  credentials_id          = dbtcloud_snowflake_credential.my_credential.credential_id
  extended_attributes_id  = dbtcloud_extended_attributes.my_attributes.extended_attributes_id
}
