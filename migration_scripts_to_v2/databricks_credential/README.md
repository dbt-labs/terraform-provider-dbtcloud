# Migration: dbtcloud_databricks_credential

## What changed

- **`resource "dbtcloud_databricks_credential"`**:
  - `adapter_type` removed — the credential type is already implicit in the resource name.
  - `target_name` removed — this field had no effect on the Databricks connection.
- **`data "dbtcloud_databricks_credential"`**:
  - `target_name` removed.

## What the script does

- Removes `adapter_type = ...` lines from all `resource "dbtcloud_databricks_credential"` blocks.
- Removes `target_name = ...` lines from all `resource "dbtcloud_databricks_credential"` blocks.
- Removes `target_name = ...` lines from all `data "dbtcloud_databricks_credential"` blocks.

## Usage

```bash
# Preview changes without modifying files
python migrate_databricks_credential.py --dry-run ./path/to/terraform/

# Apply changes (creates .tf.bak backups)
python migrate_databricks_credential.py ./path/to/terraform/

# Multiple paths
python migrate_databricks_credential.py ./envs/prod/ ./envs/staging/ ./modules/
```

## Example

**Input** (`example_input.tf`):
```hcl
resource "dbtcloud_databricks_credential" "prod_credential" {
  project_id    = 1234
  adapter_type  = "databricks"
  target_name   = "prod"
  token         = var.databricks_token
  schema        = "my_schema"
  catalog       = "my_catalog"
}

resource "dbtcloud_databricks_credential" "dev_credential" {
  project_id    = 1234
  adapter_type  = "spark"
  target_name   = "dev"
  token         = var.databricks_token_dev
  schema        = "dev_schema"
}

data "dbtcloud_databricks_credential" "existing" {
  project_id    = 1234
  credential_id = 5678
  target_name   = "prod"
}
```

**Output** (dry-run):
```
databricks_credential/example_input.tf
  [REMOVE] 2 `adapter_type` attribute(s) from resource "dbtcloud_databricks_credential" block(s)
  [REMOVE] 2 `target_name` attribute(s) from resource "dbtcloud_databricks_credential" block(s)
  [REMOVE] 1 `target_name` attribute(s) from data "dbtcloud_databricks_credential" block(s)
  (dry-run: no changes written)

Done. 1/1 file(s) had changes.
```

**Result:** `adapter_type` and `target_name` are removed from all resource blocks; `target_name` is removed from the data source block.

## Next steps

After running the script:

1. Run `terraform plan` — it should show no changes related to credential attributes.
