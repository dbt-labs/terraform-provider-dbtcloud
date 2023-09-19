# Import using a project ID and extended attribute ID found in the URL or via the API.
terraform import dbtcloud_extended_attributes.test_extended_attributes "project_id_id:extended_attributes_id"
terraform import dbtcloud_extended_attributes.test_extended_attributes 12345:6789
