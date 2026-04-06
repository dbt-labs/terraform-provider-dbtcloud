# Migration: dbtcloud_spark_credential

## What changed

- **`resource "dbtcloud_spark_credential"`**: the `target_name` attribute has been removed.
- **`data "dbtcloud_spark_credential"`**: the `target_name` attribute has been removed.

## What the script does

- Removes `target_name = ...` lines from all `resource "dbtcloud_spark_credential"` blocks.
- Removes `target_name = ...` lines from all `data "dbtcloud_spark_credential"` blocks.

## Usage

```bash
# Preview changes without modifying files
python migrate_spark_credential.py --dry-run ./path/to/terraform/

# Apply changes (creates .tf.bak backups)
python migrate_spark_credential.py ./path/to/terraform/

# Multiple paths
python migrate_spark_credential.py ./envs/prod/ ./envs/staging/ ./modules/
```

## Example

**Input** (`example_input.tf`):
```hcl
resource "dbtcloud_spark_credential" "prod_credential" {
  project_id  = 1234
  target_name = "prod"
  token       = var.spark_token
  schema      = "my_schema"
}

resource "dbtcloud_spark_credential" "dev_credential" {
  project_id  = 1234
  target_name = "dev"
  token       = var.spark_token_dev
  schema      = "dev_schema"
}

data "dbtcloud_spark_credential" "existing" {
  project_id    = 1234
  credential_id = 5678
  target_name   = "prod"
}
```

**Output** (dry-run):
```
spark_credential/example_input.tf
  [REMOVE] 2 `target_name` attribute(s) from resource "dbtcloud_spark_credential" block(s)
  [REMOVE] 1 `target_name` attribute(s) from data "dbtcloud_spark_credential" block(s)
  (dry-run: no changes written)

Done. 1/1 file(s) had changes.
```

**Result:** `target_name` is removed from all resource and data source blocks.

## Next steps

After running the script:

1. Run `terraform plan` — it should show no changes related to credential attributes.
