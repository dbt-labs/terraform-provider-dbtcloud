// use dbt_cloud_privatelink_endpoint instead of dbtcloud_privatelink_endpoint for the legacy resource names
// legacy names will be removed from 0.3 onwards

data "dbtcloud_privatelink_endpoint" "test_with_name" {
  name = "My Endpoint Name"
}

data "dbtcloud_privatelink_endpoint" "test_with_url" {
  private_link_endpoint_url = "abc.privatelink.def.com"

}
// in case multiple endpoints have the same name or URL
data "dbtcloud_privatelink_endpoint" "test_with_name_and_url" {
  name = "My Endpoint Name"
  private_link_endpoint_url = "abc.privatelink.def.com"
}
