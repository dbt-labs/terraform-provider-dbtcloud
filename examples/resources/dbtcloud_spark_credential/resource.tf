resource "dbtcloud_spark_credential" "my_spark_cred" {
  project_id = dbtcloud_project.dbt_project.id
  token      = "abcdefgh"
  schema     = "my_schema"
}

