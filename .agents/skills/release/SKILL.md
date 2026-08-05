---
name: release
description: Prepare or publish a safe gocron release by choosing a SemVer bump, updating AppVersion and migrations, invoking the complete verification gate, committing, checking CI, and creating an annotated tag. Use when asked to release, 发版, bump the version, cut a release, 打 tag, or assess release readiness.
---

# Release gocron

Keep preparation separate from publication. A version-bump request authorizes
the local version edit and verification, but not a push, merge, or tag unless
the user also asks to publish the release.

## Compatibility evidence (required)

Before tagging, record evidence for every affected case:

- `N` frontend/client with `N-1` and `N` backend;
- `N` backend with `N-1` and `N` frontend/client;
- `N` gocron with `N-1` and `N` gocron-node, in both directions;
- `N` upgrading an existing SQLite, MySQL, and PostgreSQL database;
- rolling upgrade with temporary `N`/`N-1` HA members when HA is affected.

Use automated compatibility tests, previous-version fixtures, or a documented
reproducible check. Current/current tests and a fresh install are not enough.
If a case cannot remain compatible, stop before tagging: obtain explicit
approval, use a major version, and document migration, deployment order,
rollback limits, and recovery.

## Performance evidence (required when affected)

- Identify affected hot paths and record the workload, data size, environment,
  command, baseline, and release-candidate result.
- Compare latency, throughput, memory/allocations, goroutines, query count,
  database writes, and migration duration as applicable.
- Include SQLite concurrent read/write and lock behavior when database write
  frequency, logging, polling, migrations, or scheduler state changed.
- Require a benchmark or regression test for a material fix or risk. Do not
  publish a material unexplained regression; an accepted tradeoff needs the
  user's explicit approval and release-note disclosure.

## 1. Establish the release

- Inspect the current branch, working tree, latest `v*` tag, `AppVersion`, and
  changes since the latest release.
- Never develop or commit directly on `master`; create or use a release branch.
- Choose the SemVer bump from user-facing changes: features bump minor and
  fixes or chores bump patch. Ask when the intended version cannot be inferred
  safely. Honor an explicit version supplied by the user.
- Check that the target version and `vX.Y.Z` tag do not already exist.

## 2. Update version and migrations

- Update `AppVersion` in `cmd/gocron/gocron.go`.
- If the release changes the schema, add the model to the `Install` tables
  slice where applicable, append a unique migration version id and matching
  `upgradeForNNN` function in `internal/models/migration.go`, add it to the
  upgrade chain, and add a migration test.
- Derive the migration id using the repository's existing version conversion
  convention. Confirm it with `ToNumberVersion`; do not guess or reuse an id.
- If there is no schema change, do not add an empty migration.
- For migrations that rewrite existing data, tell the user to back up the
  database and include the impact in release notes.

## 3. Pass the mandatory gate

Invoke `$verify` and require every backend, frontend, lint, test, and Docker
check to pass. Do not commit, merge, push, or tag while any check is failed,
skipped, or unavailable.

Review the final diff and confirm `AppVersion`, migration id, tests, and
user-facing release notes agree.

## 4. Publish only when requested

- Commit with `chore(release): bump version to X.Y.Z`; keep the subject under
  100 characters and omit `Co-Authored-By`.
- Push the branch and merge it into `master` according to the repository's
  normal review process.
- Confirm the GitHub Actions CI run for the exact release commit is green.
- Confirm local `master`, remote `master`, and the green CI SHA all identify the
  same commit.
- Create and push an annotated tag:
  `git tag -a vX.Y.Z -m "vX.Y.Z"` followed by
  `git push origin vX.Y.Z`.
- Confirm the tag-triggered release workflow starts successfully.

Never move or replace an existing release tag, and never tag an unverified
commit.
