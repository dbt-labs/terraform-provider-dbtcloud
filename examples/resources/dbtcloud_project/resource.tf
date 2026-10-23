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

# the ID of a global connection the project uses for analytics (it needs the analytics credentials feature)
resource "dbtcloud_project" "dbt_project_with_analytics_connection" {
  name                    = "Analytics with an analytics connection"
  analytics_connection_id = 123
}
