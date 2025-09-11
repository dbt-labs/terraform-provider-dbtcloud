terraform {
  required_providers {
    dbtcloud = {
      source  = "dbt-labs/dbtcloud"
      version = "~> 1.0"
    }
  }
}
provider "dbtcloud" {
  account_id = 70403103957642
  token      = "dbtu_nyPFnTwdKjpQ1y-u3d0ZQlc-pJ1Tr2_sHx_n5woCfK4YLqHtoc"
  host_url   = "https://wj335.us1.dbt.com/api"
}

# Create Credential
resource "dbtcloud_snowflake_credential" "terraform_test_dev_credential" {
  project_id  = 70403103992857
  auth_type   = "keypair"
  num_threads = 16
  schema      = "dummy"
  role 		  = "terraform"
  database    = "dummy"
  user = "test"
  password = "password"
  warehouse = "dummy"
}

# Create Environment
resource "dbtcloud_environment" "terraform_test_dev_env" {
  dbt_version     = "versionless"
  name            = "Dev"
  project_id      = 70403103992857
  type            = "deployment"
  connection_id   = "70403103948407"
  credential_id   = dbtcloud_snowflake_credential.terraform_test_dev_credential.credential_id
}

# Create Envrionment Variable
resource "dbtcloud_environment_variable" "dbt_terraform_env_var" {
  name       = "DBT_TERRAFORM_ENV_VAR"
  project_id = 70403103992857
  environment_values = {
    "Dev" : "testing_dev_env_value",
  }
  depends_on = [
    dbtcloud_environment.terraform_test_dev_env,
  ]
}

# Create Job
resource "dbtcloud_job" "terraform_job" {
  environment_id = dbtcloud_environment.terraform_test_dev_env.environment_id
  execute_steps = [
    "dbt debug"
  ]
  generate_docs        = false
  is_active            = false
  name                 = "Terraform Job Example"
  num_threads          = 64
  project_id           = 70403103992857
  run_generate_sources = false
  target_name          = "default"
  triggers = {
    "github_webhook" : false
    "git_provider_webhook" : false
    "schedule" : true
    "on_merge" : false
  }
  # this is the default that gets set up when modifying jobs in the UI
  schedule_days  = [0, 1, 2, 3, 4, 5, 6]
  schedule_type  = "days_of_week"
  schedule_hours = [0]
}