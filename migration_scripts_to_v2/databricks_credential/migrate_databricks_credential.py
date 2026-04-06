#!/usr/bin/env python3
"""
Migration script: dbtcloud_databricks_credential changes (v1 -> v2)

Changes in v2:
  - resource "dbtcloud_databricks_credential":
      - `adapter_type` attribute removed (was deprecated; use the credential type itself)
      - `target_name` attribute removed (was deprecated; no replacement needed)
  - data "dbtcloud_databricks_credential":
      - `target_name` attribute removed

This script removes the above attributes from all matching blocks.

Usage:
    python migrate_databricks_credential.py [--dry-run] <path> [<path> ...]

    <path> can be a .tf file or a directory (searched recursively).

Examples:
    python migrate_databricks_credential.py ./terraform/
    python migrate_databricks_credential.py --dry-run ./envs/prod/ ./envs/staging/
    python migrate_databricks_credential.py module1/main.tf module2/main.tf
"""

import argparse
import re
import shutil
import sys
from pathlib import Path


def find_tf_files(paths: list[str]) -> list[Path]:
    result = []
    for p in paths:
        path = Path(p)
        if path.is_dir():
            result.extend(sorted(path.rglob("*.tf")))
        elif path.is_file() and path.suffix == ".tf":
            result.append(path)
        else:
            print(f"WARNING: skipping {p} (not a .tf file or directory)", file=sys.stderr)
    return result


def remove_attribute_from_blocks(content: str, block_type: str, resource_type: str, attr: str) -> tuple[str, int]:
    """Remove `attr = ...` lines that appear inside blocks of the given type and resource type."""
    block_header = re.compile(
        r'^([ \t]*)(?:' + re.escape(block_type) + r')\s+"' + re.escape(resource_type) + r'"\s+"[^"]+"\s*\{',
        re.MULTILINE,
    )
    attr_line = re.compile(r'^[ \t]*' + re.escape(attr) + r'\s*=\s*[^\n]+\n?', re.MULTILINE)

    count = 0
    result = []
    pos = 0

    for m in block_header.finditer(content):
        result.append(content[pos : m.start()])
        depth = 1
        i = m.end()
        while i < len(content) and depth > 0:
            if content[i] == "{":
                depth += 1
            elif content[i] == "}":
                depth -= 1
            i += 1
        block_content = content[m.start() : i]
        new_block, n = attr_line.subn("", block_content)
        count += n
        result.append(new_block)
        pos = i

    result.append(content[pos:])
    return "".join(result), count


def process_file(path: Path, dry_run: bool) -> bool:
    original = path.read_text(encoding="utf-8")
    content = original
    changes = []

    # Resource: remove adapter_type and target_name
    content, n = remove_attribute_from_blocks(content, "resource", "dbtcloud_databricks_credential", "adapter_type")
    if n:
        changes.append(f"  [REMOVE] {n} `adapter_type` attribute(s) from resource \"dbtcloud_databricks_credential\" block(s)")

    content, n = remove_attribute_from_blocks(content, "resource", "dbtcloud_databricks_credential", "target_name")
    if n:
        changes.append(f"  [REMOVE] {n} `target_name` attribute(s) from resource \"dbtcloud_databricks_credential\" block(s)")

    # Data source: remove target_name
    content, n = remove_attribute_from_blocks(content, "data", "dbtcloud_databricks_credential", "target_name")
    if n:
        changes.append(f"  [REMOVE] {n} `target_name` attribute(s) from data \"dbtcloud_databricks_credential\" block(s)")

    if not changes:
        return False

    print(f"\n{path}")
    for c in changes:
        print(c)

    if dry_run:
        print("  (dry-run: no changes written)")
        return True

    shutil.copy2(path, path.with_suffix(".tf.bak"))
    path.write_text(content, encoding="utf-8")
    print(f"  -> written (backup: {path.with_suffix('.tf.bak')})")
    return True


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("paths", nargs="+", help=".tf files or directories to process")
    parser.add_argument("--dry-run", action="store_true", help="show changes without writing files")
    args = parser.parse_args()

    files = find_tf_files(args.paths)
    if not files:
        print("No .tf files found.", file=sys.stderr)
        sys.exit(1)

    changed = 0
    for f in files:
        if process_file(f, args.dry_run):
            changed += 1

    print(f"\nDone. {changed}/{len(files)} file(s) had changes.")
    if changed and not args.dry_run:
        print("\nNext steps:")
        print("  1. Review the changes above.")
        print("  2. Run: terraform plan")


if __name__ == "__main__":
    main()
