resource "dbtcloud_athena_credential" "example" {
  project_id           = dbtcloud_project.example.id
  aws_access_key_id    = "your-access-key-id"
  aws_secret_access_key = "your-secret-access-key"
  schema               = "your_schema"
} 