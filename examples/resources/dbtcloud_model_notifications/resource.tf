resource "dbtcloud_model_notifications" "prod_model_notifications" {
  environment_id = dbtcloud_environment.prod_environment.environment_id
  enabled        = true
  on_success     = false
  on_failure     = true
  on_warning     = true
} 