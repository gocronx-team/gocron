# gocron

## Project Overview

A lightweight, distributed scheduled task management system written in Go with a Vue.js web interface.

## Tech Stack

- **Backend:** Go 1.26, Gin, GORM (MySQL/PostgreSQL/SQLite)
- **Frontend:** Vue 3 (Options API), Element Plus, vue-i18n, Vite, pnpm
- **RPC:** gRPC + Protocol Buffers
- **Auth:** JWT + TOTP 2FA

## Development

```bash
# Backend (with hot reload)
air

# Frontend dev server
cd web/gocronx-admin && pnpm dev

# Build frontend
cd web/gocronx-admin && pnpm build

# Run tests
go test ./...

# Build
go build ./...
```

## Pre-commit / release checks (IMPORTANT)

CI (`.github/workflows/ci.yml`) does more than `go test` — **lint is a separate
gate**. Passing `go build` / `go test` locally is NOT enough (this has bitten us:
a staticcheck issue passed tests but failed CI). Before committing or releasing,
run the full set — or just run the `/verify` command:

```bash
gofmt -l .                 # must print nothing
go vet ./...
golangci-lint run ./...    # v2.12.2, same version as CI
go test -race ./...        # CI runs tests with -race
cd web/gocronx-admin && pnpm build-only && pnpm exec vue-tsc --noEmit && pnpm lint
```

The commit-msg hook (commitlint) rejects subject lines longer than 100 chars —
use a short subject + a body for multi-point commits.

## Project Structure

```
cmd/gocron/          - Main entry point
internal/
  models/            - GORM data models
  routers/           - Gin HTTP handlers (grouped by domain)
  service/           - Business logic (scheduler, execution)
  modules/           - Utilities (logger, i18n, notify, RPC)
web/gocronx-admin/   - Vue 3 + TypeScript frontend (art-design-pro based)
  src/api/           - API client services
  src/views/         - Page components
  src/components/    - Shared components
  src/locales/langs/ - i18n (zh.json, en.json)
  src/router/        - Vue Router config
  src/store/         - Pinia stores
```

## Conventions

- Commit messages follow Conventional Commits: `feat:`, `fix:`, `chore:`, `refactor:`, `style:`, `test:`
- Do not add `Co-Authored-By` lines in commit messages
- Backend i18n: `internal/modules/i18n/zh_cn.go` and `en_us.go`
- Frontend i18n: `web/gocronx-admin/src/locales/langs/zh.json` and `en.json`
- Database migrations: `internal/models/migration.go`. To add a table/column:
  (1) add the model to the `tables` slice in `Install`; (2) append a new
  `versionId` + `upgradeForNNN` to the `Upgrade` chain; (3) the id tracks
  `AppVersion` (e.g. v1.7.0 → 170); (4) add a migration test. Never reuse ids.
- **Versioning:** `AppVersion` lives in `cmd/gocron/gocron.go`. Features bump
  minor (1.6.x → 1.7.0), fixes bump patch. Do NOT tag a release until `/verify`
  is fully green — see the `release` skill for the flow.
- Do not develop directly on `master`; work on a branch and merge.
