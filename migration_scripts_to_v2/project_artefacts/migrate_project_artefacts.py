#!/usr/bin/env python3
"""
Migration script: dbtcloud_project_artefacts removal (v1 -> v2)

The dbtcloud_project_artefacts resource has been removed in v2.
This script finds all usages and removes them from your Terraform configs.

IMPORTANT: After running this script, also remove the resource from state:
    terraform state rm dbtcloud_project_artefacts.<name>

Usage:
    python migrate_project_artefacts.py [--dry-run] <path> [<path> ...]

    <path> can be a .tf file or a directory (searched recursively).

Examples:
    python migrate_project_artefacts.py ./terraform/
    python migrate_project_artefacts.py --dry-run ./envs/prod/ ./envs/staging/
    python migrate_project_artefacts.py module1/main.tf module2/main.tf
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


def remove_resource_block(content: str, resource_type: str) -> tuple[str, list[str]]:
    """
    Remove all blocks matching:
        resource "<resource_type>" "<name>" { ... }

    Handles nested braces. Returns (new_content, list_of_removed_block_names).
    """
    removed = []
    pattern = re.compile(
        r'^([ \t]*)resource\s+"' + re.escape(resource_type) + r'"\s+"([^"]+)"\s*\{',
        re.MULTILINE,
    )

    result = []
    pos = 0
    for m in pattern.finditer(content):
        result.append(content[pos : m.start()])
        name = m.group(2)
        # walk forward to find matching closing brace
        depth = 1
        i = m.end()
        while i < len(content) and depth > 0:
            if content[i] == "{":
                depth += 1
            elif content[i] == "}":
                depth -= 1
            i += 1
        # consume trailing newline(s)
        while i < len(content) and content[i] == "\n":
            i += 1
        removed.append(name)
        pos = i

    result.append(content[pos:])
    return "".join(result), removed


def process_file(path: Path, dry_run: bool) -> bool:
    original = path.read_text(encoding="utf-8")
    new_content, removed = remove_resource_block(original, "dbtcloud_project_artefacts")

    if not removed:
        return False

    print(f"\n{path}")
    for name in removed:
        print(f"  [REMOVE] resource \"dbtcloud_project_artefacts\" \"{name}\"")
        print(f"           -> also run: terraform state rm dbtcloud_project_artefacts.{name}")

    if dry_run:
        print("  (dry-run: no changes written)")
        return True

    shutil.copy2(path, path.with_suffix(".tf.bak"))
    path.write_text(new_content, encoding="utf-8")
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
        print("  2. Remove each resource from state:")
        print("       terraform state rm dbtcloud_project_artefacts.<name>")
        print("  3. Run: terraform plan")


if __name__ == "__main__":
    main()
