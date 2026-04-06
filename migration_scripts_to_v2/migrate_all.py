#!/usr/bin/env python3
"""
Master migration script: dbt Cloud Terraform Provider v1 -> v2

Runs all individual migration scripts in sequence against the supplied paths.

Usage:
    python migrate_all.py [--dry-run] <path> [<path> ...]

    <path> can be a .tf file or a directory (searched recursively).

Examples:
    python migrate_all.py ./terraform/
    python migrate_all.py --dry-run ./envs/prod/ ./envs/staging/
    python migrate_all.py module1/ module2/ module3/
"""

import argparse
import subprocess
import sys
from pathlib import Path


SCRIPTS = [
    ("project_artefacts/migrate_project_artefacts.py", "dbtcloud_project_artefacts (resource removed)"),
    ("webhook/migrate_webhook.py",                     "dbtcloud_webhook (webhook_id removed)"),
    ("repository/migrate_repository.py",               "dbtcloud_repository (fetch_deploy_key removed)"),
    ("databricks_credential/migrate_databricks_credential.py", "dbtcloud_databricks_credential (adapter_type, target_name removed)"),
    ("spark_credential/migrate_spark_credential.py",   "dbtcloud_spark_credential (target_name removed)"),
    ("job/migrate_job.py",                             "dbtcloud_job (timeout_seconds migrated to execution block; data source attrs removed)"),
]


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("paths", nargs="+", help=".tf files or directories to process")
    parser.add_argument("--dry-run", action="store_true", help="show changes without writing files")
    args = parser.parse_args()

    scripts_dir = Path(__file__).parent
    python = sys.executable

    print("=" * 70)
    print("dbt Cloud Terraform Provider: v1 -> v2 migration")
    print("=" * 70)

    for script_name, description in SCRIPTS:
        script_path = scripts_dir / script_name
        print(f"\n{'─' * 70}")
        print(f"Running: {description}")
        print(f"Script:  {script_name}")
        print(f"{'─' * 70}")

        cmd = [python, str(script_path)] + (["--dry-run"] if args.dry_run else []) + args.paths
        result = subprocess.run(cmd)
        if result.returncode not in (0, 1):
            print(f"\nERROR: {script_name} exited with code {result.returncode}", file=sys.stderr)

    print(f"\n{'=' * 70}")
    print("All migration scripts complete.")
    if args.dry_run:
        print("This was a dry run — no files were modified.")
        print("Re-run without --dry-run to apply changes.")
    else:
        print("\nNext steps:")
        print("  1. Review all [WARN] messages above for items needing manual attention.")
        print("  2. For dbtcloud_project_artefacts removals, run:")
        print("       terraform state rm dbtcloud_project_artefacts.<name>")
        print("  3. Run: terraform init -upgrade")
        print("  4. Run: terraform plan")
    print("=" * 70)


if __name__ == "__main__":
    main()
