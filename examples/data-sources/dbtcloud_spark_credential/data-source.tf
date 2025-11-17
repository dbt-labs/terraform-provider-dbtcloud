data "dbtcloud_spark_credential" "my_spark_cred" {
  project_id    = dbtcloud_project.dbt_project.id
  credential_id = 12345
}

