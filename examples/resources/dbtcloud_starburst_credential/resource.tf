resource "dbtcloud_starburst_credential" "example" {
  project_id = dbtcloud_project.example.id
  database = "your_catalog"
  schema = "your_schema"
  user = "your_user"
  password = "your_password"
}