#!/usr/bin/env python3
"""
Migration script: dbtcloud_job changes (v1 -> v2)

Changes in v2:
  - resource "dbtcloud_job":
      - Top-level `timeout_seconds` removed. Use the `execution` block instead:
            # Before
            timeout_seconds = 3600

            # After
            execution = {
              timeout_seconds = 3600
            }
        If a resource already has an `execution` block with `timeout_seconds`, the
        top-level attribute is simply removed (no duplicate is created).

  - data "dbtcloud_job" (singular):
      - `deferring_job_id` attribute removed (no replacement in data source).
      - References to `.deferring_job_id` in expressions are flagged with a warning.

  - data "dbtcloud_jobs" (plural):
      - `deferring_job_definition_id` attribute removed.

This script handles all of the above automatically where safe, and flags cases
that require manual review.

Usage:
    python migrate_job.py [--dry-run] <path> [<path> ...]

    <path> can be a .tf file or a directory (searched recursively).

Examples:
    python migrate_job.py ./terraform/
    python migrate_job.py --dry-run ./envs/prod/ ./envs/staging/
    python migrate_job.py module1/main.tf module2/main.tf
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


def find_blocks(content: str, block_type: str, resource_type: str) -> list[tuple[int, int, str]]:
    """
    Return list of (start, end, label) tuples for every matching block.
    `end` points to the character after the closing brace.
    """
    pattern = re.compile(
        r'^[ \t]*' + re.escape(block_type) + r'\s+"' + re.escape(resource_type) + r'"\s+"([^"]+)"\s*\{',
        re.MULTILINE,
    )
    blocks = []
    for m in pattern.finditer(content):
        label = m.group(1)
        depth = 1
        i = m.end()
        while i < len(content) and depth > 0:
            if content[i] == "{":
                depth += 1
            elif content[i] == "}":
                depth -= 1
            i += 1
        blocks.append((m.start(), i, label))
    return blocks


def remove_attribute_in_range(content: str, start: int, end: int, attr: str) -> tuple[str, int]:
    """Remove bare `attr = ...` lines within content[start:end]."""
    attr_re = re.compile(r'^[ \t]*' + re.escape(attr) + r'\s*=\s*[^\n]+\n?', re.MULTILINE)
    before = content[:start]
    block = content[start:end]
    after = content[end:]
    new_block, count = attr_re.subn("", block)
    return before + new_block + after, count


def has_execution_block(block_content: str) -> bool:
    """Return True if the block already contains an `execution` sub-block."""
    return bool(re.search(r'\bexecution\s*=?\s*\{', block_content))


def has_execution_timeout(block_content: str) -> bool:
    """Return True if an execution block already contains timeout_seconds."""
    exec_match = re.search(r'\bexecution\s*=?\s*\{', block_content)
    if not exec_match:
        return False
    # find the execution block body
    depth = 1
    i = exec_match.end()
    while i < len(block_content) and depth > 0:
        if block_content[i] == "{":
            depth += 1
        elif block_content[i] == "}":
            depth -= 1
        i += 1
    exec_body = block_content[exec_match.start():i]
    return bool(re.search(r'\btimeout_seconds\s*=', exec_body))


def migrate_job_resource_timeout(content: str) -> tuple[str, list[str]]:
    """
    For each resource "dbtcloud_job" block:
      - If `timeout_seconds = N` exists at top level AND no `execution` block exists:
        remove the top-level attribute and insert an execution block.
      - If `timeout_seconds = N` exists at top level AND `execution` block already has
        `timeout_seconds`: just remove the top-level attribute (execution wins).
      - If `timeout_seconds = N` exists at top level AND `execution` block exists but
        WITHOUT `timeout_seconds`: remove top-level and add timeout_seconds to execution.
    """
    top_level_timeout_re = re.compile(
        r'^([ \t]*)timeout_seconds\s*=\s*(\d+)\n?', re.MULTILINE
    )

    changes = []
    blocks = find_blocks(content, "resource", "dbtcloud_job")

    # Process in reverse order so offsets remain valid
    for start, end, label in reversed(blocks):
        block = content[start:end]

        m = top_level_timeout_re.search(block)
        if not m:
            continue

        indent = m.group(1)
        timeout_value = m.group(2)

        if has_execution_timeout(block):
            # execution block already has timeout_seconds — just remove top-level
            new_block = top_level_timeout_re.sub("", block, count=1)
            changes.append(
                f'  [REMOVE] top-level `timeout_seconds` from resource "dbtcloud_job" "{label}" '
                f'(execution block already has timeout_seconds)'
            )
        elif has_execution_block(block):
            # execution block exists but without timeout_seconds — add it there
            exec_re = re.compile(r'(\bexecution\s*=?\s*\{)')
            new_block = top_level_timeout_re.sub("", block, count=1)
            new_block = exec_re.sub(
                r'\1\n' + indent + '  timeout_seconds = ' + timeout_value,
                new_block,
                count=1,
            )
            changes.append(
                f'  [MIGRATE] resource "dbtcloud_job" "{label}": '
                f'moved `timeout_seconds = {timeout_value}` into existing execution block'
            )
        else:
            # No execution block at all — remove top-level and add execution block
            # Insert execution block right after removing timeout_seconds
            new_block = top_level_timeout_re.sub(
                indent + 'execution = {\n' + indent + '  timeout_seconds = ' + timeout_value + '\n' + indent + '}\n',
                block,
                count=1,
            )
            changes.append(
                f'  [MIGRATE] resource "dbtcloud_job" "{label}": '
                f'converted `timeout_seconds = {timeout_value}` to execution block'
            )

        content = content[:start] + new_block + content[end:]

    return content, changes


def remove_attribute_from_blocks(content: str, block_type: str, resource_type: str, attr: str) -> tuple[str, list[str]]:
    """Remove `attr = ...` lines from all matching blocks."""
    attr_re = re.compile(r'^[ \t]*' + re.escape(attr) + r'\s*=\s*[^\n]+\n?', re.MULTILINE)
    changes = []
    blocks = find_blocks(content, block_type, resource_type)

    for start, end, label in reversed(blocks):
        block = content[start:end]
        new_block, n = attr_re.subn("", block)
        if n:
            changes.append(
                f'  [REMOVE] {n} `{attr}` attribute(s) from {block_type} "{resource_type}" "{label}"'
            )
            content = content[:start] + new_block + content[end:]

    return content, changes


def warn_attribute_references(content: str, resource_type: str, attr: str) -> list[str]:
    """Warn about expression references to a removed attribute."""
    pattern = re.compile(
        r'(data\.' + re.escape(resource_type) + r'\.[a-zA-Z0-9_\-]+\.' + re.escape(attr) + r')'
    )
    warnings = []
    for m in pattern.finditer(content):
        line_num = content[:m.start()].count("\n") + 1
        warnings.append(
            f'  [WARN] line {line_num}: reference to removed attribute `{m.group(1)}` — '
            f'this attribute no longer exists in v2; remove or replace this reference manually.'
        )
    return warnings


def process_file(path: Path, dry_run: bool) -> bool:
    original = path.read_text(encoding="utf-8")
    content = original
    all_changes = []
    has_warnings = False

    # 1. Migrate top-level timeout_seconds in resource blocks
    content, changes = migrate_job_resource_timeout(content)
    all_changes.extend(changes)

    # 2. Remove deferring_job_id from data "dbtcloud_job" blocks
    content, changes = remove_attribute_from_blocks(content, "data", "dbtcloud_job", "deferring_job_id")
    all_changes.extend(changes)

    # 3. Remove deferring_job_definition_id from data "dbtcloud_jobs" blocks
    content, changes = remove_attribute_from_blocks(content, "data", "dbtcloud_jobs", "deferring_job_definition_id")
    all_changes.extend(changes)

    # 4. Warn about references to removed data source attributes
    warnings = warn_attribute_references(content, "dbtcloud_job", "deferring_job_id")
    if warnings:
        has_warnings = True
        all_changes.extend(warnings)

    if not all_changes:
        return False

    print(f"\n{path}")
    for c in all_changes:
        print(c)

    if dry_run:
        print("  (dry-run: no changes written)")
        return True

    if has_warnings and not all_changes[0].startswith("  [WARN]"):
        # changes to write (warnings are just informational)
        pass

    # Only write if there are actual changes (not just warnings)
    actual_changes = [c for c in all_changes if not c.startswith("  [WARN]")]
    if not actual_changes:
        return True  # warnings only, no file changes

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

    print(f"\nDone. {changed}/{len(files)} file(s) had changes or warnings.")
    if changed and not args.dry_run:
        print("\nNext steps:")
        print("  1. Review the changes above.")
        print("  2. Manually fix any [WARN] references to removed attributes.")
        print("  3. Run: terraform plan")


if __name__ == "__main__":
    main()
