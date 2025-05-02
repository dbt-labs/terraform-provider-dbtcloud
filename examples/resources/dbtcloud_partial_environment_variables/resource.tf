resource "dbtcloud_project" "test_project" {
  name = "testproject"
}

resource "dbtcloud_environment" "test_env" {
  name        = "development"
  type = "development"
  project_id = dbtcloud_project.test_project.id
}

resource "dbtcloud_environment_variable" "test_env_var" {
  project_id = dbtcloud_project.test_project.id
  name       = "DBT_TESTVAR"
  environment_values = {
    (dbtcloud_environment.test_env.name) = "devval"
  }
}

resource "dbtcloud_environment" "test_env_prod" {
  name        = "production"
  type = "deployment"
  dbt_version = "latest"
  deployment_type = "production"
  project_id = dbtcloud_project.test_project.id
}

resource "dbtcloud_partial_environment_variable" "test_env_var_partial" {
  project_id = dbtcloud_project.test_project.id
  name       = "DBT_TESTVAR"
  environment_values = {
    (dbtcloud_environment.test_env_prod.name) = "prodval"
  }
}