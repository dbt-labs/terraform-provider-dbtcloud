---
page_title: "4. Job Attribute Conflicts and Requirements"
subcategory: ""
description: |-
  Understanding mutually exclusive attributes and prerequisites for dbt Cloud job configuration.
---

# 4. Job Attribute Conflicts and Requirements

This guide documents the attribute conflicts and prerequisites for the `dbtcloud_job` resource. These constraints are enforced by the Terraform provider and/or the dbt Cloud API.

~> **Important:** Several job attributes are mutually exclusive. Setting conflicting attributes will result in validation errors during `terraform plan` or API errors during `terraform apply`.

## Quick Reference

### Constraint Summary

```yaml
# Machine-parseable constraint definitions
constraints:
  deferral:
    group: [self_deferring, deferring_environment_id, deferring_job_id]
    rule: "exactly_one_or_none"
    note: "deferring_job_id is DEPRECATED - use deferring_environment_id"
  
  schedule:
    group: [schedule_cron, schedule_interval, schedule_hours]
    rule: "exactly_one_or_none"
  
  triggers:
    condition: "on_merge == true"
    requires: "github_webhook == false AND git_provider_webhook == false AND schedule == false"

prerequisites:
  run_compare_changes:
    requires:
      - "deferring_environment_id IS SET"
      - "environment.deployment_type IN ['staging', 'production']"
      - "Account has Advanced CI enabled in dbt Cloud settings"
  
  errors_on_lint_failure:
    requires: "run_lint == true"
  
  compare_changes_flags:
    requires: "run_compare_changes == true"
```

### Conflict Matrix

| Attribute | CONFLICTS WITH | REQUIRED BY |
|-----------|----------------|-------------|
| `self_deferring` | `deferring_environment_id`, `deferring_job_id` | - |
| `deferring_environment_id` | `self_deferring`, `deferring_job_id` | `run_compare_changes` |
| `deferring_job_id` | `self_deferring`, `deferring_environment_id` | - (DEPRECATED) |
| `schedule_cron` | `schedule_interval`, `schedule_hours` | - |
| `schedule_interval` | `schedule_cron`, `schedule_hours` | - |
| `schedule_hours` | `schedule_cron`, `schedule_interval` | - |
| `on_merge = true` | Requires all other triggers = `false` | - |
| `run_compare_changes` | - | Requires `deferring_environment_id` |
| `errors_on_lint_failure` | - | Requires `run_lint = true` |

---

## Deferral Configuration

Deferral allows a job to reference artifacts (manifest, run results) from previous job runs. There are three mutually exclusive ways to configure deferral.

### Constraints

- **MUST use exactly ONE of:** `self_deferring`, `deferring_environment_id`, or `deferring_job_id`
- **MUST NOT combine** any two deferral attributes
- `deferring_job_id` is **DEPRECATED** - use `deferring_environment_id` instead

### Decision Tree

```
Do you need deferral?
├── No → Don't set any deferral attributes
└── Yes → What type?
    ├── Defer to SAME job's previous run
    │   └── Set: self_deferring = true
    │   └── MUST NOT set: deferring_environment_id, deferring_job_id
    │
    ├── Defer to another ENVIRONMENT (Recommended)
    │   └── Set: deferring_environment_id = <environment_id>
    │   └── MUST NOT set: self_deferring, deferring_job_id
    │
    └── Defer to specific JOB (DEPRECATED)
        └── Set: deferring_job_id = <job_id>
        └── MUST NOT set: self_deferring, deferring_environment_id
```

### Valid Examples

#### Environment Deferral (Recommended)

```terraform
# CORRECT: CI job defers to production environment
resource "dbtcloud_job" "ci_job" {
  name           = "CI Job"
  project_id     = dbtcloud_project.my_project.id
  environment_id = dbtcloud_environment.ci.environment_id
  execute_steps  = ["dbt build -s state:modified+"]
  
  # Environment deferral - RECOMMENDED approach
  deferring_environment_id = dbtcloud_environment.prod.environment_id
  
  # Do NOT include self_deferring or deferring_job_id
  
  triggers = {
    github_webhook       = true
    git_provider_webhook = true
    schedule             = false
    on_merge             = false
  }
}
```

#### Self-Deferral

```terraform
# CORRECT: Job defers to its own previous run (for source freshness workflows)
resource "dbtcloud_job" "freshness_job" {
  name           = "Source Freshness Job"
  project_id     = dbtcloud_project.my_project.id
  environment_id = dbtcloud_environment.prod.environment_id
  execute_steps  = ["dbt source freshness", "dbt build -s source_status:fresher+"]
  
  # Self-deferral
  self_deferring = true
  
  # Do NOT include deferring_environment_id or deferring_job_id
  
  triggers = {
    github_webhook       = false
    git_provider_webhook = false
    schedule             = true
    on_merge             = false
  }
  schedule_type  = "every_day"
  schedule_hours = [6]
}
```

### Invalid Examples

```terraform
# WRONG: Cannot combine self_deferring with deferring_environment_id
resource "dbtcloud_job" "invalid_job" {
  name           = "Invalid Job"
  project_id     = dbtcloud_project.my_project.id
  environment_id = dbtcloud_environment.ci.environment_id
  execute_steps  = ["dbt build"]
  
  # ERROR: These two attributes conflict
  self_deferring           = true                                    # REMOVE THIS
  deferring_environment_id = dbtcloud_environment.prod.environment_id
  
  triggers = {
    github_webhook       = true
    git_provider_webhook = true
    schedule             = false
    on_merge             = false
  }
}
# API Error: "A job cannot defer to both a job and an environment"
```

---

## Schedule Configuration

Jobs can be scheduled using one of three mutually exclusive methods.

### Constraints

- **MUST use exactly ONE of:** `schedule_cron`, `schedule_interval`, or `schedule_hours`
- **MUST NOT combine** any two schedule attributes
- All require `triggers.schedule = true` to take effect

### Decision Tree

```
How should the job be scheduled?
├── Custom cron expression
│   └── Set: schedule_cron = "0 */6 * * *"
│   └── Set: schedule_type = "custom_cron"
│   └── MUST NOT set: schedule_interval, schedule_hours
│
├── Every N hours (1-23)
│   └── Set: schedule_interval = 6
│   └── Set: schedule_type = "interval_cron"
│   └── MUST NOT set: schedule_cron, schedule_hours
│
└── Specific hours of day
    └── Set: schedule_hours = [9, 17]
    └── Set: schedule_type = "days_of_week" or "every_day"
    └── MUST NOT set: schedule_cron, schedule_interval
```

### Valid Examples

#### Custom Cron Expression

```terraform
# CORRECT: Run every 6 hours using cron
resource "dbtcloud_job" "cron_job" {
  name           = "Cron Scheduled Job"
  project_id     = dbtcloud_project.my_project.id
  environment_id = dbtcloud_environment.prod.environment_id
  execute_steps  = ["dbt build"]
  
  triggers = {
    github_webhook       = false
    git_provider_webhook = false
    schedule             = true
    on_merge             = false
  }
  
  schedule_type = "custom_cron"
  schedule_cron = "0 */6 * * *"
  
  # Do NOT include schedule_interval or schedule_hours
}
```

#### Interval-Based Schedule

```terraform
# CORRECT: Run every 4 hours
resource "dbtcloud_job" "interval_job" {
  name           = "Interval Scheduled Job"
  project_id     = dbtcloud_project.my_project.id
  environment_id = dbtcloud_environment.prod.environment_id
  execute_steps  = ["dbt build"]
  
  triggers = {
    github_webhook       = false
    git_provider_webhook = false
    schedule             = true
    on_merge             = false
  }
  
  schedule_type     = "interval_cron"
  schedule_interval = 4
  schedule_days     = [0, 1, 2, 3, 4, 5, 6]
  
  # Do NOT include schedule_cron or schedule_hours
}
```

#### Specific Hours

```terraform
# CORRECT: Run at 9am and 5pm on weekdays
resource "dbtcloud_job" "hourly_job" {
  name           = "Business Hours Job"
  project_id     = dbtcloud_project.my_project.id
  environment_id = dbtcloud_environment.prod.environment_id
  execute_steps  = ["dbt build"]
  
  triggers = {
    github_webhook       = false
    git_provider_webhook = false
    schedule             = true
    on_merge             = false
  }
  
  schedule_type  = "days_of_week"
  schedule_days  = [1, 2, 3, 4, 5]  # Monday through Friday
  schedule_hours = [9, 17]
  
  # Do NOT include schedule_cron or schedule_interval
}
```

### Invalid Examples

```terraform
# WRONG: Cannot combine schedule_cron with schedule_hours
resource "dbtcloud_job" "invalid_schedule" {
  name           = "Invalid Schedule Job"
  project_id     = dbtcloud_project.my_project.id
  environment_id = dbtcloud_environment.prod.environment_id
  execute_steps  = ["dbt build"]
  
  triggers = {
    github_webhook       = false
    git_provider_webhook = false
    schedule             = true
    on_merge             = false
  }
  
  # ERROR: These attributes conflict
  schedule_cron  = "0 9 * * *"  # REMOVE THIS
  schedule_hours = [9, 17]      # OR REMOVE THIS
}
# Terraform Error: "Attribute 'schedule_cron' conflicts with 'schedule_hours'"
```

---

## Trigger Configuration

Triggers determine when a job runs. The `on_merge` trigger has special exclusivity requirements.

### Constraints

- When `on_merge = true`, **ALL other triggers MUST be `false`**
- `github_webhook` and `git_provider_webhook` typically have the same value
- Setting all triggers to `false` creates a deactivated job

### Job Type Inference

| Triggers Set | Inferred Job Type |
|--------------|-------------------|
| `github_webhook = true` and/or `git_provider_webhook = true` | CI Job |
| `schedule = true` | Deploy/Scheduled Job |
| `on_merge = true` | Merge Job |
| All `false` | Deactivated Job |

### Valid Examples

#### CI Job (PR-triggered)

```terraform
triggers = {
  github_webhook       = true
  git_provider_webhook = true
  schedule             = false
  on_merge             = false
}
```

#### Scheduled Deploy Job

```terraform
triggers = {
  github_webhook       = false
  git_provider_webhook = false
  schedule             = true
  on_merge             = false
}
```

#### Merge Job

```terraform
# CORRECT: on_merge requires all others to be false
triggers = {
  github_webhook       = false  # MUST be false
  git_provider_webhook = false  # MUST be false
  schedule             = false  # MUST be false
  on_merge             = true
}
```

### Invalid Examples

```terraform
# WRONG: on_merge cannot be combined with other triggers
triggers = {
  github_webhook       = true   # ERROR: Must be false when on_merge is true
  git_provider_webhook = true   # ERROR: Must be false when on_merge is true
  schedule             = false
  on_merge             = true
}
```

---

## Advanced CI (run_compare_changes)

The `run_compare_changes` attribute enables Advanced CI features for comparing data changes.

### Prerequisites Checklist

To use `run_compare_changes = true`, ALL of the following MUST be true:

- [ ] `deferring_environment_id` IS SET (environment deferral is required)
- [ ] Target environment has `deployment_type` = `"staging"` or `"production"`
- [ ] Advanced CI is enabled in your dbt Cloud Account Settings
- [ ] Job is a CI job (`github_webhook` or `git_provider_webhook` = `true`)

### Constraints

- **REQUIRES:** `deferring_environment_id` to be set
- **REQUIRES:** Environment `deployment_type` to be `staging` or `production`
- `compare_changes_flags` only applies when `run_compare_changes = true`

### Valid Example

```terraform
# CORRECT: Advanced CI with all prerequisites met
resource "dbtcloud_job" "advanced_ci_job" {
  name           = "Advanced CI Job"
  project_id     = dbtcloud_project.my_project.id
  environment_id = dbtcloud_environment.ci.environment_id
  execute_steps  = ["dbt build -s state:modified+"]
  
  # Required: Environment deferral
  deferring_environment_id = dbtcloud_environment.prod.environment_id
  
  # Advanced CI features
  run_compare_changes   = true
  compare_changes_flags = "--select state:modified"
  
  # Must be a CI job
  triggers = {
    github_webhook       = true
    git_provider_webhook = true
    schedule             = false
    on_merge             = false
  }
}

# The deferred-to environment must have deployment_type set
resource "dbtcloud_environment" "prod" {
  name            = "Production"
  project_id      = dbtcloud_project.my_project.id
  type            = "deployment"
  deployment_type = "production"  # REQUIRED for run_compare_changes
  # ... other config
}
```

### Invalid Example

```terraform
# WRONG: run_compare_changes without deferring_environment_id
resource "dbtcloud_job" "invalid_ci_job" {
  name           = "Invalid CI Job"
  project_id     = dbtcloud_project.my_project.id
  environment_id = dbtcloud_environment.ci.environment_id
  execute_steps  = ["dbt build"]
  
  # ERROR: Missing required deferring_environment_id
  run_compare_changes = true
  
  triggers = {
    github_webhook       = true
    git_provider_webhook = true
    schedule             = false
    on_merge             = false
  }
}
# API Error: "Environment deferral is required to enable compare changes"
```

---

## Lint Configuration

The `errors_on_lint_failure` attribute depends on `run_lint`.

### Constraints

- `errors_on_lint_failure` **ONLY applies when** `run_lint = true`
- If `run_lint = false`, `errors_on_lint_failure` is ignored

### Valid Example

```terraform
# CORRECT: Lint configuration
resource "dbtcloud_job" "lint_job" {
  name           = "CI with Linting"
  project_id     = dbtcloud_project.my_project.id
  environment_id = dbtcloud_environment.ci.environment_id
  execute_steps  = ["dbt build -s state:modified+"]
  
  # Lint settings (errors_on_lint_failure requires run_lint)
  run_lint               = true
  errors_on_lint_failure = true  # Only effective because run_lint = true
  
  triggers = {
    github_webhook       = true
    git_provider_webhook = true
    schedule             = false
    on_merge             = false
  }
}
```

---

## Error Message Reference

| Error Message | Cause | Solution |
|---------------|-------|----------|
| `"A job cannot defer to both a job and an environment"` | `self_deferring` or `deferring_job_id` set together with `deferring_environment_id` | Use only ONE deferral attribute. Remove `self_deferring` when using `deferring_environment_id` |
| `"Environment deferral is required to enable compare changes"` | `run_compare_changes = true` without `deferring_environment_id` | Set `deferring_environment_id` to enable Advanced CI |
| `"Attribute 'X' conflicts with 'Y'"` | Two conflicting attributes set | Use only ONE of the conflicting group (see Conflict Matrix) |
| `"on_merge cannot be true when other triggers are enabled"` | `on_merge = true` with other triggers also `true` | Set all other triggers to `false` when using `on_merge` |

---

## Migration from Deprecated Attributes

### deferring_job_id to deferring_environment_id

The `deferring_job_id` attribute is deprecated. Migrate to environment-based deferral:

**Before (Deprecated):**
```terraform
resource "dbtcloud_job" "ci_job" {
  # ...
  deferring_job_id = dbtcloud_job.prod_job.id  # DEPRECATED
}
```

**After (Recommended):**
```terraform
resource "dbtcloud_job" "ci_job" {
  # ...
  deferring_environment_id = dbtcloud_environment.prod.environment_id
}
```

-> Environment-based deferral is more flexible and is the recommended approach for all new configurations.
