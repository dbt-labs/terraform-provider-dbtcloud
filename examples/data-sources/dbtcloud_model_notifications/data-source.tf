data "dbtcloud_model_notifications" "prod_model_notifications" {
  environment_id = dbtcloud_environment.prod_environment.environment_id
}

data "dbtcloud_model_notifications" "qa_model_notifications" {
  environment_id = 12345
}
