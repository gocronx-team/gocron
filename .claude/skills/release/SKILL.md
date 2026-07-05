---
name: release
description: Cut a gocron release — bump AppVersion, align the migration version id, run the full CI check locally, and tag only when everything is green. Use when asked to "release", "发版", "bump version", "cut a release", or "打 tag".
---

# Releasing gocron

Follow these steps in order. **Never tag before the full check is green** — a
release that fails CI is worse than a delayed one.

## 1. Decide the version (SemVer)

- `AppVersion` lives in `cmd/gocron/gocron.go`.
- **Feature** → bump minor (`1.6.7` → `1.7.0`). **Fix/chore only** → bump patch
  (`1.6.7` → `1.6.8`). Matches the project's history (v1.5.9 → v1.6.0 was a minor).
- If unsure whether the batch is "feature" or "fix", ask the user.

## 2. Bump version + align migration

- Edit `AppVersion` in `cmd/gocron/gocron.go`.
- `ToNumberVersion` maps the string to an int (`1.7.0` → `170`). If this release
  adds a DB migration, its `versionId` + `upgradeForNNN` in
  `internal/models/migration.go` MUST use that same number (e.g. `170`), and the
  `Upgrade` chain must include it. Add a migration test.
- If there is NO schema change, only `AppVersion` changes.

## 3. Run the full CI check locally (mandatory gate)

Run the `/verify` command (or the steps in CLAUDE.md → "Pre-commit / release
checks"): gofmt, go vet, **golangci-lint**, `go test -race`, and the frontend
build/typecheck/lint. `go build`/`go test` alone are NOT sufficient — lint is a
separate gate. Fix everything until all steps are green.

## 4. Commit, merge, tag

- Commit the version bump: `chore(release): bump version to X.Y.Z` (short subject;
  commitlint rejects > 100 chars).
- Merge the release branch into `master` (fast-forward preferred).
- Confirm GitHub Actions is green on the pushed commit **before** tagging.
- Tag and push: `git tag -a vX.Y.Z -m "vX.Y.Z" && git push origin vX.Y.Z`
  (annotated tag). `release.yml` listens on `v*` tags to build/publish artifacts.

## 5. If a release migration changes existing data

Migrations that rewrite existing rows (e.g. remapping `notify_status`) can't be
undone — remind the user to back up the database before upgrading, and note the
change in the release notes.

## Guardrails

- Do not tag on a red or unverified CI.
- Do not develop/commit directly on `master`; use a branch.
- Do not add `Co-Authored-By` lines to commits.
