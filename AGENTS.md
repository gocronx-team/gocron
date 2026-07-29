# AGENTS.md

Guidance for AI coding tools (Cursor, Codex, Copilot, etc.) in this repo.
Claude Code has the full version in `CLAUDE.md`; this mirrors the essentials.

**gocron** — a lightweight distributed cron/task scheduler. Backend: Go 1.26
(Gin, GORM for MySQL/PostgreSQL/SQLite, gRPC). Frontend: Vue 3 + Element Plus +
Vite under `web/gocronx-admin`.

## Commands

```bash
air                                # backend, hot reload
cd web/gocronx-admin && pnpm dev   # frontend dev server
go build ./...                     # build backend
go test -race ./...                # backend tests (CI uses -race)
```

## Before you commit or release (IMPORTANT)

CI (`.github/workflows/ci.yml`) does more than `go test` — **lint is a separate
gate**. Passing `go build`/`go test` locally is NOT enough. Run all of these
(Claude Code users: run `/verify`; Codex users: invoke `$verify`):

```bash
gofmt -l .                 # must print nothing
go vet ./...
golangci-lint run ./...    # v2.12.2, same as CI
go test -race ./...
cd web/gocronx-admin && pnpm build-only && pnpm exec vue-tsc --noEmit && pnpm lint
```

## Conventions

- Conventional Commits (`feat:` `fix:` `chore:` `refactor:` `style:` `test:`).
  No `Co-Authored-By` lines. commit-msg (commitlint) rejects subjects > 100 chars.
- **Versioning:** `AppVersion` in `cmd/gocron/gocron.go`; features → minor
  (1.6.x → 1.7.0), fixes → patch. Don't tag until CI is green.
- **Migrations** (`internal/models/migration.go`): add the model to the `Install`
  tables slice + a new `versionId`/`upgradeForNNN` (id tracks AppVersion, e.g.
  v1.7.0 → 170) + a migration test. Never reuse version ids.
- **i18n:** backend `internal/modules/i18n/{zh_cn,en_us}.go`; frontend
  `web/gocronx-admin/src/locales/langs/{zh,en}.json` — keep both languages in sync.
- Do not develop directly on `master`; branch and merge.

## Repository workflows

- Codex project skills:
  - `$verify`: reproduce every CI gate before commit, merge, or release
  - `$release`: prepare and publish a version safely
  - `$migration`: implement and validate cross-database schema/data migrations
  - `$api-feature`: keep routes, permissions, clients, types, i18n, docs, and tests aligned
  - `$scheduler-change`: protect execution semantics and race-sensitive scheduler state
  - `$security-check`: audit dependencies, trust boundaries, secrets, and exposed inputs
  - `$dependency-update`: assess and verify Dependabot or manual dependency upgrades
- Claude Code: `.claude/commands/verify.md` and `.claude/skills/release`

Use the narrow project skill while implementing or reviewing a change, then
use `$verify` as the final gate. `$release` requires that gate before tagging.

More project detail lives in `CLAUDE.md`.
