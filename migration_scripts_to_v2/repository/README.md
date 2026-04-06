# Migration: dbtcloud_repository

## What changed

- **`resource "dbtcloud_repository"`**: the `fetch_deploy_key` attribute has been removed. This attribute had no effect on the repository configuration and was only used to fetch the deploy key value.
- **`data "dbtcloud_repository"`**: the `fetch_deploy_key` attribute has been removed.

## What the script does

- Removes `fetch_deploy_key = ...` lines from all `resource "dbtcloud_repository"` blocks.
- Removes `fetch_deploy_key = ...` lines from all `data "dbtcloud_repository"` blocks.

## Usage

```bash
# Preview changes without modifying files
python migrate_repository.py --dry-run ./path/to/terraform/

# Apply changes (creates .tf.bak backups)
python migrate_repository.py ./path/to/terraform/

# Multiple paths
python migrate_repository.py ./envs/prod/ ./envs/staging/ ./modules/
```

## Example

**Input** (`example_input.tf`):
```hcl
resource "dbtcloud_repository" "prod_repo" {
  project_id         = 1234
  remote_url         = "https://github.com/my-org/my-repo.git"
  git_clone_strategy = "github_app"
  fetch_deploy_key   = false
}

resource "dbtcloud_repository" "staging_repo" {
  project_id         = 2345
  remote_url         = "git@github.com:my-org/my-repo.git"
  git_clone_strategy = "deploy_key"
  fetch_deploy_key   = true
}

data "dbtcloud_repository" "existing_repo" {
  project_id       = 1234
  repository_id    = 5678
  fetch_deploy_key = false
}
```

**Output** (dry-run):
```
repository/example_input.tf
  [REMOVE] 2 `fetch_deploy_key` attribute(s) from resource "dbtcloud_repository" block(s)
  [REMOVE] 1 `fetch_deploy_key` attribute(s) from data "dbtcloud_repository" block(s)
  (dry-run: no changes written)

Done. 1/1 file(s) had changes.
```

**Result:** `fetch_deploy_key` is removed from all resource and data source blocks.

## Next steps

After running the script:

1. Run `terraform plan` — it should show no changes related to repository attributes.
