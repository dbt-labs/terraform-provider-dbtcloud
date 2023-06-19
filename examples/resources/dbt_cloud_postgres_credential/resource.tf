// use dbt_cloud_postgres_credential instead of dbtcloud_postgres_credential for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_postgres_credential" "test_credential" {
    is_active = true
    project_id = dbt_cloud_project.test_project.id
    type = "postgres"
    default_schema = "%s"
    username = "%s"
    password = "%s"
    num_threads = 3
}