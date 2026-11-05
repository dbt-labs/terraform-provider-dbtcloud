resource "dbtcloud_project" "dbt_project" {
  name = "Analytics"
}
resource "dbtcloud_project" "dbt_project_with_description" {
  name = "Analytics with description"
  description = "My awesome analytics project"
}

resource "dbtcloud_project" "dbt_project_with_subdir" {
  name                     = "Analytics in Subdir"
  dbt_project_subdirectory = "path"
}

# a global connection the project uses for analytics
resource "dbtcloud_global_connection" "analytics" {
  name = "Analytics connection"
  snowflake = {
    account   = "my-snowflake-account"
    database  = "ANALYTICS"
    warehouse = "ANALYTICS_WH"
    allow_sso = false
  }
}

resource "dbtcloud_project" "dbt_project_with_analytics_connection" {
  name                    = "Analytics with an analytics connection"
  analytics_connection_id = dbtcloud_global_connection.analytics.id
}
