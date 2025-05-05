resource "dbtcloud_teradata_credential" "test" {
  project_id           = dbtcloud_project.example.id
  schema               = "your_schema"
  user                 = "your_user"
  password             = "your_password"
}