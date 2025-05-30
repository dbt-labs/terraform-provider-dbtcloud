---
page_title: "dbtcloud_job Resource - dbtcloud"
subcategory: ""
description: |-
  Managed a dbt Cloud job.
---

# dbtcloud_job (Resource)

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
# a job that has github_webhook and git_provider_webhook 
# set to false will be categorized as a "Deploy Job"
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
  # this is the default that gets set up when modifying jobs in the UI
  schedule_days  = [0, 1, 2, 3, 4, 5, 6]
  schedule_type  = "days_of_week"
  schedule_hours = [0]
}


# a job that has github_webhook and git_provider_webhook set 
# to true will be categorized as a "Continuous Integration Job"
resource "dbtcloud_job" "ci_job" {
  environment_id = dbtcloud_environment.ci_environment.environment_id
  execute_steps = [
    "dbt build -s state:modified+ --fail-fast"
  ]
  generate_docs            = false
  deferring_environment_id = dbtcloud_environment.prod_environment.environment_id
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

- `compare_changes_flags` (String) The model selector for checking changes in the compare changes Advanced CI feature
- `dbt_version` (String) Version number of dbt to use in this job, usually in the format 1.2.0-latest rather than core versions
- `deferring_environment_id` (Number) Environment identifier that this job defers to (new deferring approach)
- `deferring_job_id` (Number) Job identifier that this job defers to (legacy deferring approach)
- `description` (String) Description for the job
- `errors_on_lint_failure` (Boolean) Whether the CI job should fail when a lint error is found. Only used when `run_lint` is set to `true`. Defaults to `true`.
- `generate_docs` (Boolean) Flag for whether the job should generate documentation
- `is_active` (Boolean) Should always be set to true as setting it to false is the same as creating a job in a deleted state. To create/keep a job in a 'deactivated' state, check  the `triggers` config.
- `job_completion_trigger_condition` (Block List) Which other job should trigger this job when it finishes, and on which conditions (sometimes referred as 'job chaining'). (see [below for nested schema](#nestedblock--job_completion_trigger_condition))
- `job_type` (String) Can be used to enforce the job type betwen `ci`, `merge` and `scheduled`. Without this value the job type is inferred from the triggers configured
- `num_threads` (Number) Number of threads to use in the job
- `run_compare_changes` (Boolean) Whether the CI job should compare data changes introduced by the code changes. Requires `deferring_environment_id` to be set. (Advanced CI needs to be activated in the dbt Cloud Account Settings first as well)
- `run_generate_sources` (Boolean) Flag for whether the job should add a `dbt source freshness` step to the job. The difference between manually adding a step with `dbt source freshness` in the job steps or using this flag is that with this flag, a failed freshness will still allow the following steps to run.
- `run_lint` (Boolean) Whether the CI job should lint SQL changes. Defaults to `false`.
- `schedule_cron` (String) Custom cron expression for schedule
- `schedule_days` (List of Number) List of days of week as numbers (0 = Sunday, 7 = Saturday) to execute the job at if running on a schedule
- `schedule_hours` (List of Number) List of hours to execute the job at if running on a schedule
- `schedule_interval` (Number) Number of hours between job executions if running on a schedule
- `schedule_type` (String) Type of schedule to use, one of every_day/ days_of_week/ custom_cron/ interval_cron
- `self_deferring` (Boolean) Whether this job defers on a previous run of itself
- `target_name` (String) Target name for the dbt profile
- `timeout_seconds` (Number, Deprecated) [Deprectated - Moved to execution.timeout_seconds] Number of seconds to allow the job to run before timing out
- `triggers_on_draft_pr` (Boolean) Whether the CI job should be automatically triggered on draft PRs

### Read-Only

- `id` (Number) The ID of this resource
- `job_id` (Number) Job identifier

<a id="nestedatt--triggers"></a>
### Nested Schema for `triggers`

Optional:

- `git_provider_webhook` (Boolean) Whether the job runs automatically on PR creation
- `github_webhook` (Boolean) Whether the job runs automatically on PR creation
- `on_merge` (Boolean) Whether the job runs automatically once a PR is merged
- `schedule` (Boolean) Whether the job runs on a schedule


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