terraform {
  required_providers {
    dbtcloud = {
      source = "dbt-labs/dbtcloud"
    }
  }
}
resource "dbtcloud_project" "dbt_project" {
  name = "My Cool Test project for repository"
}

resource "dbtcloud_project" "dbt_project2" {
  name = "My Cool Other project for repository"
}

// 1. Create connections and credentials for all credential types

# snowflake
resource "dbtcloud_global_connection" "snowflake" {
  name = "My Cool Snowflake connection"

  snowflake = {
    account                   = "my-snowflake-account"
    database                  = "MY_DATABASE"
    warehouse                 = "MY_WAREHOUSE"
    client_session_keep_alive = false
    allow_sso                 = true
    oauth_client_id           = "yourclientid"
    oauth_client_secret       = "yourclientsecret"
  }
}
resource "dbtcloud_snowflake_credential" "snowflake_credential" {
  project_id  = dbtcloud_project.dbt_project.id
  auth_type   = "password"
  num_threads = 16
  schema      = "SCHEMA"
  user        = "user"
  password    = "password"
}
resource "dbtcloud_environment" "snowflake_env" {
  name        = "My Cool Snowflake Dev Environment"
  project_id  = dbtcloud_project.dbt_project.id
  type = "deployment"
  dbt_version = "latest"
  connection_id = dbtcloud_global_connection.snowflake.id
  credential_id = dbtcloud_snowflake_credential.snowflake_credential.credential_id
}
resource "dbtcloud_job" "snowflake_daily_job" {
  environment_id = dbtcloud_environment.snowflake_env.environment_id
  execute_steps = [
    "dbt build"
  ]
  generate_docs        = true
  is_active            = true
  name                 = "Daily job"
  num_threads          = 64
  project_id           = dbtcloud_project.dbt_project.id
  run_generate_sources = true
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
resource "dbtcloud_job" "snowflake_ci_job" {
  environment_id = dbtcloud_environment.snowflake_env.environment_id
  execute_steps = [
    "dbt build -s state:modified+ --fail-fast"
  ]
  generate_docs            = false
  deferring_environment_id = dbtcloud_environment.snowflake_env.environment_id
  name                     = "CI Job"
  num_threads              = 32
  project_id               = dbtcloud_project.dbt_project.id
  run_generate_sources     = false
  run_lint                 = true
  errors_on_lint_failure   = true
  triggers = {
    "github_webhook" : true
    "git_provider_webhook" : true
    "schedule" : false
    "on_merge" : false
  }
  # this is the default that gets set up when modifying jobs in the UI
  # this is not going to be used when schedule is set to false
  schedule_days = [0, 1, 2, 3, 4, 5, 6]
  schedule_type = "days_of_week"
}
resource "dbtcloud_job" "snowflake_downstream_job" {
  environment_id = dbtcloud_environment.snowflake_env.environment_id
  execute_steps = [
    "dbt build -s +my_model"
  ]
  generate_docs            = true
  name                     = "Downstream job in project 2"
  num_threads              = 32
  project_id               = dbtcloud_project.dbt_project.id
  run_generate_sources     = true
  triggers = {
    "github_webhook" : false
    "git_provider_webhook" : false
    "schedule" : false
    "on_merge" : false
  }
  schedule_days = [0, 1, 2, 3, 4, 5, 6]
  schedule_type = "days_of_week"
  job_completion_trigger_condition {
    job_id = dbtcloud_job.snowflake_daily_job.id
    project_id = dbtcloud_project.dbt_project.id
    statuses = ["success"]
  }
}

# bigquery
resource "dbtcloud_global_connection" "bigquery" {
  name = "My Cool BigQuery connection"
  bigquery = {
    gcp_project_id              = "my-gcp-project-id"
    timeout_seconds             = 1000
    private_key_id              = "my-private-key-id"
    private_key                 = "ABCDEFGHIJKL"
    client_email                = "my_client_email"
    client_id                   = "my_client_id"
    auth_uri                    = "my_auth_uri"
    token_uri                   = "my_token_uri"
    auth_provider_x509_cert_url = "my_auth_provider_x509_cert_url"
    client_x509_cert_url        = "my_client_x509_cert_url"
    application_id              = "oauth_application_id"
    application_secret          = "oauth_secret_id"
  }
}
resource "dbtcloud_bigquery_credential" "bigquery_credential" {
  project_id  = dbtcloud_project.dbt_project.id
  dataset     = "my_bq_dataset"
  num_threads = 16
}
resource "dbtcloud_environment" "bq_env" {
  name        = "My Cool BigQuery Dev Environment"
  project_id  = dbtcloud_project.dbt_project.id
  type = "deployment"
  dbt_version = "latest"
  connection_id = dbtcloud_global_connection.bigquery.id
  credential_id = dbtcloud_bigquery_credential.bigquery_credential.credential_id
}

# redshift
resource "dbtcloud_global_connection" "redshift" {
  name = "My Cool Redshift connection"
  redshift = {
    hostname = "my-redshift-connection.com"
    port     = 5432
    // optional fields
    dbname = "my_database"
    // it is possible to set settings to connect via SSH Tunnel as well
  }
}
resource "dbtcloud_postgres_credential" "redshift_credential" {
  is_active      = true
  project_id     = dbtcloud_project.dbt_project.id
  type           = "redshift"
  default_schema = "my_schema"
  username       = "my_username"
  password       = "my_password"
  num_threads    = 16
}
resource "dbtcloud_environment" "redshift_env" {
  name        = "My Cool Redshift Dev Environment"
  project_id  = dbtcloud_project.dbt_project.id
  type = "deployment"
  dbt_version = "latest"
  connection_id = dbtcloud_global_connection.redshift.id
  credential_id = dbtcloud_postgres_credential.redshift_credential.credential_id
}

# apache_spark
resource "dbtcloud_global_connection" "apache_spark" {
  name = "My Cool Apache Spark connection"
  apache_spark = {
    method  = "http"
    host    = "my-spark-host.com"
    cluster = "my-cluster"
    // example of optional fields
    connect_timeout = 100
  }
}
resource "dbtcloud_athena_credential" "apache_spark_credential" {
  project_id           = dbtcloud_project.dbt_project.id
  aws_access_key_id    = "your-access-key-id"
  aws_secret_access_key = "your-secret-access-key"
  schema               = "your_schema"
}
resource "dbtcloud_environment" "apache_spark_env" {
  name        = "My Cool Apache Spark Dev Environment"
  project_id  = dbtcloud_project.dbt_project.id
  type = "deployment"
  dbt_version = "latest"
  connection_id = dbtcloud_global_connection.apache_spark.id
  credential_id = dbtcloud_athena_credential.apache_spark_credential.credential_id
} 

# athena
resource "dbtcloud_global_connection" "athena" {
  name = "My Cool Athena connection"
  athena = {
    region_name    = "us-east-1"
    database       = "mydatabase"
    s3_staging_dir = "my_dir"
    // example of optional fields
    work_group = "my_work_group"
  }
}
resource "dbtcloud_athena_credential" "athena_credential" {
  project_id           = dbtcloud_project.dbt_project.id
  aws_access_key_id    = "your-access-key-id"
  aws_secret_access_key = "your-secret-access-key"
  schema               = "your_schema"
}
resource "dbtcloud_environment" "athena_env" {
  name        = "My Cool Athena Dev Environment"
  project_id  = dbtcloud_project.dbt_project.id
  type = "deployment"
  dbt_version = "latest"
  connection_id = dbtcloud_global_connection.athena.id
  credential_id = dbtcloud_athena_credential.athena_credential.credential_id
}

# databricks
resource "dbtcloud_global_connection" "databricks" {
  name = "My Cool Databricks connection"
  databricks = {
    host      = "my-databricks-host.cloud.databricks.com"
    http_path = "/sql/my/http/path"
    // optional fields
    catalog       = "dbt_catalog"
    client_id     = "yourclientid"
    client_secret = "yourclientsecret"
  }
}
resource "dbtcloud_databricks_credential" "databricks_credential" {
  project_id   = dbtcloud_project.dbt_project.id
  token        = "abcdefgh"
  schema       = "my_schema"
  adapter_type = "databricks"
}
resource "dbtcloud_environment" "databricks_env" {
  name        = "My Cool Databricks Dev Environment"
  project_id  = dbtcloud_project.dbt_project.id
  type = "deployment"
  dbt_version = "latest"
  connection_id = dbtcloud_global_connection.databricks.id
  credential_id = dbtcloud_databricks_credential.databricks_credential.credential_id
}

# fabric
resource "dbtcloud_global_connection" "fabric" {
  name = "My Cool Fabric connection"
  fabric = {
    server   = "my-fabric-server.com"
    database = "mydb"
    // optional fields
    port          = 1234
    retries       = 3
    login_timeout = 60
    query_timeout = 3600
  }
}
resource "dbtcloud_fabric_credential" "fabric_credential" {
  project_id           = dbtcloud_project.dbt_project.id
  schema               = "my_schema"
  user                 = "my_user"
  password             = "my_password"
  schema_authorization = "abcd"
  adapter_type         = "fabric"

}
resource "dbtcloud_environment" "fabric_env" {
  name        = "My Cool Fabric Dev Environment"
    project_id  = dbtcloud_project.dbt_project.id
    type = "deployment"
    dbt_version = "latest"
    connection_id = dbtcloud_global_connection.fabric.id
    credential_id = dbtcloud_fabric_credential.fabric_credential.credential_id
}

# fabric with Service Principal
resource "dbtcloud_fabric_credential" "fabric_credential_serv_princ" {
  project_id           = dbtcloud_project.dbt_project.id
  schema               = "my_schema"
  client_id            = "my_client_id"
  tenant_id            = "my_tenant_id"
  client_secret        = "my_secret"
  schema_authorization = "abcd"
  adapter_type         = "fabric"
}
resource "dbtcloud_environment" "fabric_env_serv_princ" {
  name        = "My Cool Fabric Dev Environment (Service Principal)"
    project_id  = dbtcloud_project.dbt_project.id
    type = "deployment"
    dbt_version = "latest"
    connection_id = dbtcloud_global_connection.fabric.id
    credential_id = dbtcloud_fabric_credential.fabric_credential_serv_princ.credential_id
}

# postgres
resource "dbtcloud_global_connection" "postgres" {
  name = "My Cool PostgreSQL connection"
  postgres = {
    hostname = "my-postgresql-server.com"
    port     = 5432
    // optional fields
    dbname = "my_database"
    // it is possible to set settings to connect via SSH Tunnel as well
  }
}
resource "dbtcloud_postgres_credential" "postgres_credential" {
  is_active      = true
  project_id     = dbtcloud_project.dbt_project.id
  type           = "postgres"
  default_schema = "my_schema"
  username       = "my_username"
  password       = "my_password"
  num_threads    = 16
}
resource "dbtcloud_environment" "postgres_env" {
  name        = "My Cool PostgreSQL Dev Environment"
  project_id  = dbtcloud_project.dbt_project.id
    type = "deployment"
    dbt_version = "latest"
    connection_id = dbtcloud_global_connection.postgres.id
    credential_id = dbtcloud_postgres_credential.postgres_credential.credential_id
}

# starburst
resource "dbtcloud_global_connection" "starburst" {
  name = "My Cool Starburst connection"
  starburst = {
    host     = "my-starburst-host.com"
    database = "mydb"
  }
}
resource "dbtcloud_starburst_credential" "starburst_credential" {
  project_id = dbtcloud_project.dbt_project.id
  database = "your_catalog"
  schema = "your_schema"
  user = "your_user"
  password = "your_password"
}
resource "dbtcloud_environment" "starburst_env" {
  name        = "My Cool Starburst Dev Environment"
  project_id  = dbtcloud_project.dbt_project.id
  type = "deployment"
  dbt_version = "latest"
  connection_id = dbtcloud_global_connection.starburst.id
  credential_id = dbtcloud_starburst_credential.starburst_credential.credential_id
}

# synapse
resource "dbtcloud_global_connection" "synapse" {
  name = "My Cool Synapse connection"
  synapse = {
    host     = "my-synapse-server.com"
    database = "mydb"
    // optional fields
    port          = 1234
    retries       = 3
    login_timeout = 60
    query_timeout = 3600
  }
}
resource "dbtcloud_fabric_credential" "synapse_credential" {
  project_id           = dbtcloud_project.dbt_project.id
  schema               = "my_schema"
  user                 = "my_user"
  password             = "my_password"
  schema_authorization = "abcd"
  adapter_type         = "fabric"

}
resource "dbtcloud_environment" "synapse_env" {
  name        = "My Cool Synapse Dev Environment"
  project_id  = dbtcloud_project.dbt_project.id
  type = "deployment"
  dbt_version = "latest"
  connection_id = dbtcloud_global_connection.synapse.id
  credential_id = dbtcloud_fabric_credential.synapse_credential.credential_id
}

# teradata
resource "dbtcloud_global_connection" "teradata" {
  name = "My Cool Teradata connection"

  teradata = {
    host       = "teradata.example.com"
    port       		= 1025
    tmode	   		= "ANSI"
    retries	   		= 3
    request_timeout = 3000
  }
}
resource "dbtcloud_teradata_credential" "teradata_credential" {
  project_id           = dbtcloud_project.dbt_project.id
  schema               = "your_schema"
  user                 = "your_user"
  password             = "your_password"
}
resource "dbtcloud_environment" "teradata_env" {
  name        = "My Cool Teradata Dev Environment"
  project_id  = dbtcloud_project.dbt_project.id
    type = "deployment"
    dbt_version = "latest"
    connection_id = dbtcloud_global_connection.teradata.id
    credential_id = dbtcloud_teradata_credential.teradata_credential.credential_id
}