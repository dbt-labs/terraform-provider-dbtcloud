# Migration: dbtcloud_webhook

## What changed

- **`resource "dbtcloud_webhook"`**: the `webhook_id` attribute has been removed. Use the `id` attribute instead — it holds the same value and has always been populated.
- **`data "dbtcloud_webhook"`**: the `webhook_id` attribute has been removed. The `id` attribute is now `Required` (you must supply it to look up a webhook).

## What the script does

- Removes `webhook_id = ...` lines from all `resource "dbtcloud_webhook"` blocks.
- Removes `webhook_id = ...` lines from all `data "dbtcloud_webhook"` blocks.
- Rewrites any expression references of the form `data.dbtcloud_webhook.<name>.webhook_id` to `data.dbtcloud_webhook.<name>.id`.

## Usage

```bash
# Preview changes without modifying files
python migrate_webhook.py --dry-run ./path/to/terraform/

# Apply changes (creates .tf.bak backups)
python migrate_webhook.py ./path/to/terraform/

# Multiple paths
python migrate_webhook.py ./envs/prod/ ./envs/staging/ ./modules/
```

## Example

**Input** (`example_input.tf`):
```hcl
resource "dbtcloud_webhook" "job_completed" {
  client_url  = "https://example.com/webhook"
  event_types = ["job.run.completed"]
  name        = "Job Completed Hook"
  webhook_id  = "wsu_abc123"
}

resource "dbtcloud_webhook" "job_started" {
  client_url  = "https://example.com/webhook/start"
  event_types = ["job.run.started"]
  name        = "Job Started Hook"
  webhook_id  = "wsu_def456"
}

data "dbtcloud_webhook" "existing_hook" {
  id         = "wsu_abc123"
  webhook_id = "wsu_abc123"
}

output "webhook_endpoint" {
  value = data.dbtcloud_webhook.existing_hook.webhook_id
}
```

**Output** (dry-run):
```
webhook/example_input.tf
  [REMOVE] 2 `webhook_id` attribute(s) from resource "dbtcloud_webhook" block(s)
  [REMOVE] 1 `webhook_id` attribute(s) from data "dbtcloud_webhook" block(s)
  [REPLACE] 1 reference(s) `.webhook_id` -> `.id` in expressions
  (dry-run: no changes written)

Done. 1/1 file(s) had changes.
```

**Result:** `webhook_id` is removed from all resource and data source blocks, and the output reference is rewritten to use `.id`.

## Next steps

After running the script:

1. Check any `data "dbtcloud_webhook"` blocks to confirm `id` is set — this is now a required field.
2. Run `terraform plan` — it should show no changes related to webhook attributes.
