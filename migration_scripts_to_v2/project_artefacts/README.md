# Migration: dbtcloud_project_artefacts

## What changed

The `dbtcloud_project_artefacts` resource has been **removed** in v2. There is no replacement — the project artefacts configuration is no longer managed as a separate resource.

## What the script does

- Finds every `resource "dbtcloud_project_artefacts"` block in your `.tf` files and removes it entirely.
- Prints the name of each removed block along with the `terraform state rm` command you must run manually to clean up the state.

> **Note:** The script only modifies your configuration files. You must remove the resources from your Terraform state separately (see Next Steps below).

## Usage

```bash
# Preview changes without modifying files
python migrate_project_artefacts.py --dry-run ./path/to/terraform/

# Apply changes (creates .tf.bak backups)
python migrate_project_artefacts.py ./path/to/terraform/

# Multiple paths
python migrate_project_artefacts.py ./envs/prod/ ./envs/staging/ ./modules/
```

## Example

**Input** (`example_input.tf`):
```hcl
resource "dbtcloud_project_artefacts" "prod_artefacts" {
  project_id       = 1234
  docs_job_id      = 5678
  freshness_job_id = 9012
}

resource "dbtcloud_project_artefacts" "staging_artefacts" {
  project_id  = 2345
  docs_job_id = 6789
}
```

**Output** (dry-run):
```
project_artefacts/example_input.tf
  [REMOVE] resource "dbtcloud_project_artefacts" "prod_artefacts"
           -> also run: terraform state rm dbtcloud_project_artefacts.prod_artefacts
  [REMOVE] resource "dbtcloud_project_artefacts" "staging_artefacts"
           -> also run: terraform state rm dbtcloud_project_artefacts.staging_artefacts
  (dry-run: no changes written)

Done. 1/1 file(s) had changes.
```

**Result:** Both resource blocks are deleted from the file.

## Next steps

After running the script:

1. For each removed resource, remove it from your Terraform state:
   ```bash
   terraform state rm dbtcloud_project_artefacts.<name>
   ```
2. Run `terraform init -upgrade` to pick up the new provider version.
3. Run `terraform plan` — it should show no changes related to `dbtcloud_project_artefacts`.
