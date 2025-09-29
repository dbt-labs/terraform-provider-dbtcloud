// use dbt_cloud_privatelink_endpoints instead of dbtcloud_privatelink_endpoints for the legacy resource names
// legacy names will be removed from 0.3 onwards

data "dbtcloud_privatelink_endpoints" "all" {}

# Find a specific endpoint by name
locals {
  snowflake_endpoint = [
    for endpoint in data.dbtcloud_privatelink_endpoints.all.endpoints :
    endpoint if endpoint.name == "Snowflake Production Endpoint"
  ][0]
}

# Use the endpoint in a global connection
resource "dbtcloud_global_connection" "snowflake" {
  name                     = "Snowflake via PrivateLink"
  private_link_endpoint_id = local.snowflake_endpoint.id

  snowflake = {
    account   = "my-snowflake-account"
    database  = "ANALYTICS"
    warehouse = "COMPUTE_WH"
  }
}

# Filter endpoints by type
locals {
  snowflake_endpoints = [
    for endpoint in data.dbtcloud_privatelink_endpoints.all.endpoints : 
    endpoint if endpoint.type == "snowflake"
  ]
}

# Create connections for all Snowflake endpoints
resource "dbtcloud_global_connection" "snowflake_connections" {
  for_each = { for ep in local.snowflake_endpoints : ep.id => ep }

  name                     = "Connection for ${each.value.name}"
  private_link_endpoint_id = each.value.id

  snowflake = {
    account   = "my-account"
    database  = "ANALYTICS"
    warehouse = "COMPUTE_WH"
  }
}
