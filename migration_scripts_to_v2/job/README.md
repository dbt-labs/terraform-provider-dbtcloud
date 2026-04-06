# Migration: dbtcloud_job

## What changed

- **`resource "dbtcloud_job"`**: the top-level `timeout_seconds` attribute has been removed. Set execution timeout via the `execution` block instead:
  ```hcl
  # Before
  resource "dbtcloud_job" "my_job" {
    timeout_seconds = 3600
  }

  # After
  resource "dbtcloud_job" "my_job" {
    execution = {
      timeout_seconds = 3600
    }
  }
  ```

- **`data "dbtcloud_job"` (singular)**: the `deferring_job_id` attribute has been removed from the data source schema. It is still available on the resource.

- **`data "dbtcloud_jobs"` (plural)**: the `deferring_job_definition_id` attribute has been removed.

## What the script does

**For `resource "dbtcloud_job"` blocks**, the script handles three cases:

| Situation | Action |
|---|---|
| `timeout_seconds` at top level, no `execution` block | Creates an `execution` block with `timeout_seconds` |
| `timeout_seconds` at top level + `execution` block without `timeout_seconds` | Moves `timeout_seconds` into the existing `execution` block |
| `timeout_seconds` at top level + `execution` block already has `timeout_seconds` | Removes the top-level attribute (the `execution` block value takes precedence) |

**For `data "dbtcloud_job"` blocks**: removes `deferring_job_id = ...` lines.

**For `data "dbtcloud_jobs"` blocks**: removes `deferring_job_definition_id = ...` lines.

**Expression references**: any reference to `data.dbtcloud_job.<name>.deferring_job_id` in expressions is flagged with a `[WARN]` — these require manual cleanup as there is no direct replacement.

## Usage

```bash
# Preview changes without modifying files
python migrate_job.py --dry-run ./path/to/terraform/

# Apply changes (creates .tf.bak backups)
python migrate_job.py ./path/to/terraform/

# Multiple paths
python migrate_job.py ./envs/prod/ ./envs/staging/ ./modules/
```

## Example

**Input** (`example_input.tf`):
```hcl
# Case 1: top-level timeout_seconds only
resource "dbtcloud_job" "deploy_job" {
  ...
  timeout_seconds = 3600
}

# Case 2: top-level timeout_seconds + execution block without timeout
resource "dbtcloud_job" "ci_job" {
  ...
  timeout_seconds = 1800
  execution = {
  }
}

# Case 3: both top-level and execution.timeout_seconds set
resource "dbtcloud_job" "another_job" {
  ...
  timeout_seconds = 7200
  execution = {
    timeout_seconds = 900
  }
}

data "dbtcloud_job" "my_job" {
  job_id           = 1234
  deferring_job_id = 5678
}

data "dbtcloud_jobs" "all_jobs" {
  project_id                  = 5678
  deferring_job_definition_id = 1234
}
```

**Output** (dry-run):
```
job/example_input.tf
  [REMOVE]  top-level `timeout_seconds` from resource "dbtcloud_job" "another_job" (execution block already has timeout_seconds)
  [MIGRATE] resource "dbtcloud_job" "ci_job": moved `timeout_seconds = 1800` into existing execution block
  [MIGRATE] resource "dbtcloud_job" "deploy_job": converted `timeout_seconds = 3600` to execution block
  [REMOVE]  1 `deferring_job_id` attribute(s) from data "dbtcloud_job" "my_job"
  [REMOVE]  1 `deferring_job_definition_id` attribute(s) from data "dbtcloud_jobs" "all_jobs"
  (dry-run: no changes written)

Done. 1/1 file(s) had changes or warnings.
```

## Next steps

After running the script:

1. Manually fix any `[WARN]` references to `data.dbtcloud_job.<name>.deferring_job_id` — this attribute no longer exists on the data source.
2. Run `terraform plan` — it should show no changes related to job attributes.
