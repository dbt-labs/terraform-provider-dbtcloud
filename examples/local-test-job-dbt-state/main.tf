terraform {
  required_providers {
    dbtcloud = {
      source = "dbt-labs/dbtcloud"
    }
  }
}

# Auth is read from the environment so no secrets live in this file:
#   export DBT_CLOUD_ACCOUNT_ID=...
#   export DBT_CLOUD_TOKEN=...
#   export DBT_CLOUD_HOST_URL=https://cloud.getdbt.com/api   # adjust if needed
provider "dbtcloud" {}

variable "dbt_version" {
  description = "dbt version for the environment. Use a Fusion track (e.g. latest-fusion) so both SAO and dbt State work."
  type        = string
  default     = "latest-fusion"
}

resource "dbtcloud_project" "test" {
  name = "tf-cost-opt-test"
}

resource "dbtcloud_environment" "test" {
  project_id  = dbtcloud_project.test.id
  name        = "tf-cost-opt-test-env"
  dbt_version = var.dbt_version
  type        = "deployment"
  # SAO (without the allow-all-environments flag) additionally requires a
  # production/staging deployment environment. dbt State does not.
  deployment_type = "production"
}

# ── Case A: dbt State ────────────────────────────────────────────────────────
# Requires the account to have dbt State enabled (ORC-3638-enable-dbt-state +
# the dbt State entitlement). dbt State is environment-independent.
# Expect after apply: cost_optimization_features = ["dbt_state"] and
# force_node_selection = false (derived by the API).
resource "dbtcloud_job" "dbt_state" {
  name           = "tf-job-dbt-state"
  project_id     = dbtcloud_project.test.id
  environment_id = dbtcloud_environment.test.environment_id

  execute_steps = ["dbt build"]

  triggers = {
    github_webhook       = false
    git_provider_webhook = false
    schedule             = false
    on_merge             = false
  }

  cost_optimization_features = ["dbt_state"]
}

output "dbt_state_force_node_selection" {
  description = "Should be false — derived by the API from cost_optimization_features."
  value       = dbtcloud_job.dbt_state.force_node_selection
}

output "dbt_state_cost_optimization_features" {
  value = dbtcloud_job.dbt_state.cost_optimization_features
}

# ── Case B: State-Aware Orchestration ────────────────────────────────────────
# Requires the orc-2664-sao-beta flag (and a Fusion dbt_version + prod/staging
# env, or the orc-2714-allow-sao-for-all-environments flag).
# Expect after apply: cost_optimization_features = ["state_aware_orchestration"]
# and force_node_selection = false.
# resource "dbtcloud_job" "sao" {
#   name           = "tf-job-sao"
#   project_id     = dbtcloud_project.test.id
#   environment_id = dbtcloud_environment.test.environment_id

#   execute_steps = ["dbt build"]

#   triggers = {
#     github_webhook       = false
#     git_provider_webhook = false
#     schedule             = false
#     on_merge             = false
#   }

#   cost_optimization_features = ["state_aware_orchestration"]
# }

# output "sao_force_node_selection" {
#   value = dbtcloud_job.sao.force_node_selection
# }

# ── Case C: validator rejects mixing dbt_state with other features ───────────
# This should fail at PLAN time (terraform plan) with a validation error,
# before any API call. Uncomment to verify.
# resource "dbtcloud_job" "invalid_combo" {
#   name           = "tf-job-invalid"
#   project_id     = dbtcloud_project.test.id
#   environment_id = dbtcloud_environment.test.environment_id
#   execute_steps  = ["dbt build"]
#   triggers = {
#     github_webhook       = false
#     git_provider_webhook = false
#     schedule             = false
#     on_merge             = false
#   }
#   cost_optimization_features = ["dbt_state", "state_aware_orchestration"]
# }

# ── Case D: disable everything ───────────────────────────────────────────────
# Set cost_optimization_features = [] (or remove it) and re-apply on Case A/B.
# Expect: cost_optimization_features = [] and force_node_selection = true.
