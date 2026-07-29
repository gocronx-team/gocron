---
name: migration
description: Create, review, or verify gocron database migrations across SQLite, MySQL, and PostgreSQL. Use when adding or changing GORM models, columns, indexes, constraints, persisted settings, migration version ids, Install tables, upgradeForNNN functions, or migration tests.
---

# Migrate gocron data

Treat migrations as compatibility code. Preserve existing installations and all
three supported databases; do not optimize only for a fresh SQLite database.

## Establish the change

- Inspect `cmd/gocron/gocron.go`, `internal/models/migration.go`, the affected
  models, and nearby migration tests before editing.
- Determine whether the change affects an existing table, creates a new table,
  or transforms existing data. Do not create an empty migration for a release
  with no schema or data change.
- Derive the migration id from the target `AppVersion` using the repository's
  established conversion. Never reuse an id. If the release version is not
  known and a new id is required, stop and ask for it.

## Implement atomically

For a new persisted model or schema change:

1. Add a new unique id to `versionIds` in chronological release order.
2. Add the matching `migration.upgradeForNNN` entry at the same index.
3. Implement `upgradeForNNN` with the transaction passed to it; return every
   error instead of logging and continuing.
4. Add new install-time models to the `Install` table slice. Do not add an
   existing-table column there as a substitute for an upgrade.
5. Make the upgrade idempotent where practical. Guard data rewrites that would
   corrupt values on a second run.
6. Avoid database-specific SQL. When unavoidable, branch on the GORM dialect
   and implement SQLite, MySQL, and PostgreSQL behavior explicitly.
7. Add a focused migration test that builds the pre-upgrade schema, runs the
   upgrade, checks the resulting schema/data, and exercises a second run when
   idempotency is expected.

Do not rely only on GORM `AutoMigrate` when data must be renamed, remapped,
backfilled, deduplicated, or constrained.

## Verify

Run:

```bash
python3 .agents/skills/migration/scripts/check_migration.py
go test -race ./internal/models/...
go test -race ./cmd/gocron/...
```

If a specific migration was added, pass its id to the structural checker:

```bash
python3 .agents/skills/migration/scripts/check_migration.py 191
```

Then invoke `$verify` before committing. Report the migration id, affected
tables, upgrade path, downgrade/backup implications, database-specific risks,
and exact tests run.
