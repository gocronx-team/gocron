#!/usr/bin/env python3
"""Check that gocron migration ids and handlers remain structurally aligned."""

from __future__ import annotations

import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[4]
MIGRATION = ROOT / "internal/models/migration.go"


def fail(message: str) -> None:
    print(f"FAIL: {message}")
    raise SystemExit(1)


text = MIGRATION.read_text(encoding="utf-8")

ids_match = re.search(r"versionIds\s*:=\s*\[\]int\s*\{([^}]*)\}", text, re.S)
funcs_match = re.search(
    r"upgradeFuncs\s*:=\s*\[\]func\(\*gorm\.DB\) error\s*\{([^}]*)\}", text, re.S
)
if not ids_match or not funcs_match:
    fail("cannot locate versionIds or upgradeFuncs in internal/models/migration.go")

ids = [int(value) for value in re.findall(r"\d+", ids_match.group(1))]
handlers = [int(value) for value in re.findall(r"upgradeFor(\d+)", funcs_match.group(1))]

if len(ids) != len(set(ids)):
    fail("versionIds contains a duplicate id")
if ids != handlers:
    fail(f"versionIds and upgradeFuncs differ: ids={ids}, handlers={handlers}")

definitions = [int(value) for value in re.findall(r"func\s+\([^)]*\)\s+upgradeFor(\d+)\s*\(", text)]
missing = [value for value in ids if value not in definitions]
if missing:
    fail(f"missing upgrade function definitions for: {missing}")

if len(sys.argv) > 2 or (len(sys.argv) == 2 and not sys.argv[1].isdigit()):
    fail("usage: check_migration.py [expected-version-id]")
if len(sys.argv) == 2:
    expected = int(sys.argv[1])
    if expected not in ids:
        fail(f"expected migration id {expected} is not registered")
    if not re.search(rf"upgradeFor{expected}\b", text):
        fail(f"expected upgradeFor{expected} is missing")

print(f"PASS: {len(ids)} migration ids and handlers are aligned")
