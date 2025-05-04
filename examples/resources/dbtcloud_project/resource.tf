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