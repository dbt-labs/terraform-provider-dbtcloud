---
page_title: "dbtcloud_job Resource - dbtcloud"
subcategory: ""
description: |-
  Managed a dbt Cloud job.
---

# dbtcloud_job (Resource)

~> **Attribute Conflicts:** Several job attributes are mutually exclusive. See the [Job Attribute Conflicts Guide](../guides/4_job_attribute_conflicts.md) for detailed decision trees and examples.

## Attribute Conflict Quick Reference

| Conflict Group | Attributes (use only ONE) |
|---------------|---------------------------|
| **Deferral** | `self_deferring`, `deferring_environment_id`, `deferring_job_id` |
| **Schedule** | `schedule_cron`, `schedule_interval`, `schedule_hours` |
| **Triggers** | When `on_merge = true`, all other triggers MUST be `false` |

**Prerequisites:**
- `run_compare_changes` REQUIRES `deferring_environment_id` to be set
- `errors_on_lint_failure` REQUIRES `run_lint = true`
- `compare_changes_flags` REQUIRES `run_compare_changes = true`

~> In October 2023, CI improvements have been rolled out to dbt Cloud with minor impacts to some jobs:  [more info](https://docs.getdbt.com/docs/dbt-versions/release-notes/june-2023/ci-updates-phase1-rn). 
<br/>
<br/>
Those improvements include modifications to deferral which was historically set at the job level and will now be set at the environment level. 
Deferral can still be set to "self" by setting `self_deferring` to `true` but with the new approach, deferral to other runs need to be done with `deferring_environment_id` instead of `deferring_job_id`.

~> New with 0.3.1, `triggers` now accepts a `on_merge` value to trigger jobs when code is merged in git. If `on_merge` is `true` all other triggers need to be `false`.
<br/>
<br/>
For now, it is not a mandatory field, but it will be in a future version. Please add `on_merge` in your config or modules. 

## Example Usage

```terraform
# DEPLOY JOB: github_webhook and git_provider_webhook set to false
# Uses schedule_hours - MUST NOT also set schedule_cron or schedule_interval
resource "dbtcloud_job" "daily_job" {
  environment_id = dbtcloud_environment.prod_environment.environment_id
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
  
  # SCHEDULE CONFLICT: Use only ONE of schedule_hours, schedule_interval, or schedule_cron
  schedule_days  = [0, 1, 2, 3, 4, 5, 6]
  schedule_type  = "days_of_week"
  schedule_hours = [0]
  # Do NOT add schedule_cron or schedule_interval - they conflict with schedule_hours
}


# CI JOB: github_webhook and git_provider_webhook set to true
# Uses deferring_environment_id - MUST NOT also set self_deferring or deferring_job_id
resource "dbtcloud_job" "ci_job" {
  environment_id = dbtcloud_environment.ci_environment.environment_id
  execute_steps = [
    "dbt build -s state:modified+ --fail-fast"
  ]
  generate_docs = false
  name          = "CI Job"
  num_threads   = 32
  project_id    = dbtcloud_project.dbt_project.id
  
  # DEFERRAL CONFLICT: Use only ONE of deferring_environment_id, self_deferring, or deferring_job_id
  deferring_environment_id = dbtcloud_environment.prod_environment.environment_id
  # Do NOT add self_deferring or deferring_job_id - they conflict with deferring_environment_id
  
  # LINT DEPENDENCY: errors_on_lint_failure only applies when run_lint = true
  run_lint               = true
  errors_on_lint_failure = true
  run_generate_sources   = false
  
  triggers = {
    "github_webhook" : true
    "git_provider_webhook" : true
    "schedule" : false
    "on_merge" : false
  }
  
  schedule_days = [0, 1, 2, 3, 4, 5, 6]
  schedule_type = "days_of_week"
}

# a job that is set to be triggered after another job finishes
# this is sometimes referred as 'job chaining'
resource "dbtcloud_job" "downstream_job" {
  environment_id = dbtcloud_environment.project2_prod_environment.environment_id
  execute_steps = [
    "dbt build -s +my_model"
  ]
  generate_docs            = true
  name                     = "Downstream job in project 2"
  num_threads              = 32
  project_id               = dbtcloud_project.dbt_project2.id
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
    job_id = dbtcloud_job.daily_job.id
    project_id = dbtcloud_project.dbt_project.id
    statuses = ["success"]
  }
}

# a job that uses the interval cron setup
resource "dbtcloud_job" "daily_job" {
  environment_id = dbtcloud_environment.prod_environment.environment_id
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

  schedule_type  = "interval_cron"
  schedule_days = [0, 1, 2, 3, 4, 5, 6]
  schedule_interval = 5
}
```


## Advanced Usage for setting up deferral

To use deferral in dbt Cloud, some jobs need to run successfully so that we can levereage the manifest file generated.

It is possible to have those jobs run automatically based on their schedule, or to run them manually, but in the scenarion where many dbt Cloud projects are created, it might be easier to automate this process.

### For Continuous Integration

In the case of Continuous Integration, our CI job needs to defer to the Production environment.
So, we need to have a successful run in the Production environment before the CI process can execute as expected.

The example below shows how the Terraform config can be updated to automatically trigger a run of the job in the Production environment, leveraging the `local-exec` provisioner and `curl` to trigger the run.

```tf
# a periodic job, but we trigger it once with `dbt parse` as soon as it is created so we can defer to the environment it is in
# to do so, we use a local-exec provisioner, just make sure that the machine running Terraform has curl installed
resource "dbtcloud_job" "daily_job" {
  environment_id = dbtcloud_environment.prod_environment.environment_id
  execute_steps = [
    "dbt build"
  ]
  generate_docs        = true
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

  schedule_days  = [0, 1, 2, 3, 4, 5, 6]
  schedule_type  = "days_of_week"
  schedule_hours = [0]

  provisioner "local-exec" {
    command = <<-EOT
      response=$(curl -s -L -o /dev/null -w "%%{http_code}" -X POST \
        -H 'Authorization: Bearer ${var.dbt_token}' \
        -H 'Content-Type: application/json' \
        -d '{"cause": "Generate manifest", "steps_override": ["dbt parse"]}' \
        ${var.dbt_host_url}/v2/accounts/${var.dbt_account_id}/jobs/${self.id}/run/)
      
      if [ "$response" -ge 200 ] && [ "$response" -lt 300 ]; then
        echo "Success: HTTP status $response"
        exit 0
      else
        echo "Failure: HTTP status $response"
        exit 1
      fi
    EOT
  }
}
```

### For allowing source freshness deferral

In the case that deferral is required so that we can use [the `source_status:fresher+` selector](https://docs.getdbt.com/docs/build/sources#build-models-based-on-source-freshness), 
the process is more complicated as the job will be self deferring.

An example can be found [in this GitHub issue](https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/360#issuecomment-2779336961).

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `environment_id` (Number) Environment ID to create the job in
- `execute_steps` (List of String) List of commands to execute for the job
- `name` (String) Job name
- `project_id` (Number) Project ID to create the job in
- `triggers` (Attributes) Flags for which types of triggers to use, the values are `github_webhook`, `git_provider_webhook`, `schedule` and `on_merge`. All flags should be listed and set with `true` or `false`. When `on_merge` is `true`, all the other values must be false.<br>`custom_branch_only` used to be allowed but has been deprecated from the API. The jobs will use the custom branch of the environment. Please remove the `custom_branch_only` from your config. <br>To create a job in a 'deactivated' state, set all to `false`. (see [below for nested schema](#nestedatt--triggers))

### Optional

- `compare_changes_flags` (String) The model selector for checking changes in the compare changes Advanced CI feature. **REQUIRES:** `run_compare_changes = true`
- `dbt_version` (String) Version number of dbt to use in this job, usually in the format 1.2.0-latest rather than core versions
- `deferring_environment_id` (Number) Environment identifier that this job defers to (new deferring approach). **CONFLICTS WITH:** `self_deferring`, `deferring_job_id`. **REQUIRED BY:** `run_compare_changes`
- `deferring_job_id` (Number, Deprecated) Job identifier that this job defers to. **DEPRECATED:** Use `deferring_environment_id` instead. **CONFLICTS WITH:** `self_deferring`, `deferring_environment_id`
- `description` (String) Description for the job
- `errors_on_lint_failure` (Boolean) Whether the CI job should fail when a lint error is found. **REQUIRES:** `run_lint = true`. Defaults to `true`.
- `force_node_selection` (Boolean) Whether to force node selection (SAO - Select All Optimizations) for the job. If `dbt_version` is not set to `latest-fusion`, this must be set to `true` when specified.
- `generate_docs` (Boolean) Flag for whether the job should generate documentation
- `is_active` (Boolean) Should always be set to true as setting it to false is the same as creating a job in a deleted state. To create/keep a job in a 'deactivated' state, check  the `triggers` config. Setting it to false essentially deletes the job. On resource creation, this field is enforced to be true.
- `job_completion_trigger_condition` (Block List) Which other job should trigger this job when it finishes, and on which conditions (sometimes referred as 'job chaining'). (see [below for nested schema](#nestedblock--job_completion_trigger_condition))
- `job_type` (String) Can be used to enforce the job type betwen `ci`, `merge` and `scheduled`. Without this value the job type is inferred from the triggers configured
- `num_threads` (Number) Number of threads to use in the job
- `run_compare_changes` (Boolean) Whether the CI job should compare data changes introduced by the code changes. **REQUIRES:** `deferring_environment_id` to be set AND environment `deployment_type` to be `staging` or `production`. (Advanced CI needs to be activated in the dbt Cloud Account Settings first as well)
- `run_generate_sources` (Boolean) Flag for whether the job should add a `dbt source freshness` step to the job. The difference between manually adding a step with `dbt source freshness` in the job steps or using this flag is that with this flag, a failed freshness will still allow the following steps to run.
- `run_lint` (Boolean) Whether the CI job should lint SQL changes. Defaults to `false`.
- `schedule_cron` (String) Custom cron expression for schedule. **CONFLICTS WITH:** `schedule_interval`, `schedule_hours`
- `schedule_days` (List of Number) List of days of week as numbers (0 = Sunday, 7 = Saturday) to execute the job at if running on a schedule
- `schedule_hours` (List of Number) List of hours to execute the job at if running on a schedule. **CONFLICTS WITH:** `schedule_cron`, `schedule_interval`
- `schedule_interval` (Number) Number of hours between job executions if running on a schedule. **CONFLICTS WITH:** `schedule_cron`, `schedule_hours`
- `schedule_type` (String) Type of schedule to use, one of every_day/ days_of_week/ custom_cron/ interval_cron
- `self_deferring` (Boolean) Whether this job defers on a previous run of itself. **CONFLICTS WITH:** `deferring_environment_id`, `deferring_job_id`
- `target_name` (String) Target name for the dbt profile
- `timeout_seconds` (Number, Deprecated) [Deprectated - Moved to execution.timeout_seconds] Number of seconds to allow the job to run before timing out
- `triggers_on_draft_pr` (Boolean) Whether the CI job should be automatically triggered on draft PRs
- `validate_execute_steps` (Boolean) When set to `true`, the provider will validate the `execute_steps` during plan time to ensure they contain valid dbt commands. If a command is not recognized (e.g., a new dbt command not yet supported by the provider), the validation will fail. Defaults to `false` to allow flexibility with newer dbt commands.

### Read-Only

- `id` (Number) The ID of this resource
- `job_id` (Number) Job identifier

<a id="nestedatt--triggers"></a>
### Nested Schema for `triggers`

~> **Trigger Exclusivity:** When `on_merge = true`, ALL other triggers (`github_webhook`, `git_provider_webhook`, `schedule`) MUST be set to `false`.

Optional:

- `git_provider_webhook` (Boolean) Whether the job runs automatically on PR creation. **MUST be `false` when** `on_merge = true`
- `github_webhook` (Boolean) Whether the job runs automatically on PR creation. **MUST be `false` when** `on_merge = true`
- `on_merge` (Boolean) Whether the job runs automatically once a PR is merged. **REQUIRES:** All other triggers to be `false`
- `schedule` (Boolean) Whether the job runs on a schedule. **MUST be `false` when** `on_merge = true`


<a id="nestedblock--job_completion_trigger_condition"></a>
### Nested Schema for `job_completion_trigger_condition`

Required:

- `job_id` (Number) The ID of the job that would trigger this job after completion.
- `project_id` (Number) The ID of the project where the trigger job is running in.
- `statuses` (Set of String) List of statuses to trigger the job on. Possible values are `success`, `error` and `canceled`.

## Import

Import is supported using the following syntax:

```shell
# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_job.my_job
  id = "job_id"
}

import {
  to = dbtcloud_job.my_job
  id = "12345"
}

# using the older import command
terraform import dbtcloud_job.my_job "job_id"
terraform import dbtcloud_job.my_job 12345
```