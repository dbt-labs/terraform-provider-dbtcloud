#!/usr/bin/env python3
"""
Migration script: dbtcloud_webhook changes (v1 -> v2)

Changes in v2:
  - resource "dbtcloud_webhook": the `webhook_id` attribute has been removed.
    Use the `id` attribute instead.
  - data "dbtcloud_webhook": the `webhook_id` attribute has been removed.
    The `id` attribute is now Required (you must supply it to look up a webhook).

This script:
  1. Removes `webhook_id = ...` lines from dbtcloud_webhook resource blocks.
  2. Removes `webhook_id = ...` lines from dbtcloud_webhook data source blocks.
  3. Replaces references to `.webhook_id` attribute in expressions with `.id`.

Usage:
    python migrate_webhook.py [--dry-run] <path> [<path> ...]

    <path> can be a .tf file or a directory (searched recursively).

Examples:
    python migrate_webhook.py ./terraform/
    python migrate_webhook.py --dry-run ./envs/prod/ ./envs/staging/
    python migrate_webhook.py module1/main.tf module2/main.tf
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
    # Match the block header
    block_header = re.compile(
        r'^([ \t]*)(?:' + re.escape(block_type) + r')\s+"' + re.escape(resource_type) + r'"\s+"[^"]+"\s*\{',
        re.MULTILINE,
    )
    # Match the attribute line (handles quoted strings, numbers, booleans, references)
    attr_line = re.compile(r'^[ \t]*' + re.escape(attr) + r'\s*=\s*[^\n]+\n?', re.MULTILINE)

    count = 0
    result = []
    pos = 0

    for m in block_header.finditer(content):
        result.append(content[pos : m.start()])
        # find the end of the block (matching closing brace)
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


def replace_attribute_references(content: str, resource_type: str, old_attr: str, new_attr: str) -> tuple[str, int]:
    """Replace data.<resource_type>.<name>.<old_attr> with data.<resource_type>.<name>.<new_attr>."""
    pattern = re.compile(
        r'\b(data\.' + re.escape(resource_type) + r'\.[a-zA-Z0-9_\-]+)\.' + re.escape(old_attr) + r'\b'
    )
    new_content, count = pattern.subn(r'\1.' + new_attr, content)
    return new_content, count


def process_file(path: Path, dry_run: bool) -> bool:
    original = path.read_text(encoding="utf-8")
    content = original
    changes = []

    # Remove webhook_id from resource blocks
    content, n = remove_attribute_from_blocks(content, "resource", "dbtcloud_webhook", "webhook_id")
    if n:
        changes.append(f"  [REMOVE] {n} `webhook_id` attribute(s) from resource \"dbtcloud_webhook\" block(s)")

    # Remove webhook_id from data source blocks
    content, n = remove_attribute_from_blocks(content, "data", "dbtcloud_webhook", "webhook_id")
    if n:
        changes.append(f"  [REMOVE] {n} `webhook_id` attribute(s) from data \"dbtcloud_webhook\" block(s)")

    # Replace .webhook_id references with .id in expressions
    content, n = replace_attribute_references(content, "dbtcloud_webhook", "webhook_id", "id")
    if n:
        changes.append(f"  [REPLACE] {n} reference(s) `.webhook_id` -> `.id` in expressions")

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
        print("  2. For each data \"dbtcloud_webhook\" block that previously used `webhook_id`")
        print("     to look up a webhook, make sure `id` is now explicitly set.")
        print("  3. Run: terraform plan")


if __name__ == "__main__":
    main()
