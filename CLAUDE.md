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

## Compatibility policy (MANDATORY)

Backward compatibility is the default and is not optional.

- Release `N` must interoperate with the immediately previous released minor
  version `N-1`. This applies to frontend/backend, gocron/gocron-node, and
  temporary `N`/`N-1` HA mixtures during rolling upgrades.
- Both deployment orders must work when a component boundary is affected. New
  capabilities require detection and a safe fallback when the peer is `N-1`.
- Existing installations must upgrade in place without losing tasks,
  configuration, users, or logs. Migrations must support SQLite, MySQL, and
  PostgreSQL and must not require a clean install.
- Preserve existing HTTP routes, methods, accepted requests, response fields
  and types, status codes, SSE contracts, and user-visible semantics. New
  fields must be optional and have backward-compatible defaults.
- Protobuf/gRPC evolution must be additive: never renumber or reuse fields,
  remove fields still used by a supported version, or silently change an
  existing RPC's meaning. Keep the previous path until the window closes.
- Preserve execution, retry, cancellation, and status-transition semantics
  across mixed versions. Never retry after an ambiguous remote accept.
- Test each affected boundary with previous-version schemas, fixtures,
  contracts, or protocol doubles. Current-only and fresh-install-only tests are
  insufficient evidence for that boundary.
- A breaking change requires explicit user approval before implementation, a
  major version, and migration, rollout, and rollback plans. Never hide one in
  a minor or patch release. If compatibility is uncertain, stop and ask.

## Performance policy (MANDATORY)

Material performance risk requires evidence, not intuition.

- Establish a representative baseline before materially changing scheduler hot
  paths, high-volume list/search APIs, polling or streaming, task logs, large
  database queries, queues, serialization, or populated-table migrations.
- Define the workload and data size, then compare before/after latency,
  throughput, allocations or memory, goroutine count, database query count,
  and write frequency as applicable. Report commands and results.
- Do not introduce N+1 queries, unbounded reads, responses, logs, goroutines,
  queues, retries, polling, caches, or in-memory accumulation. Require limits,
  pagination/batching, cancellation, cleanup, and backpressure where relevant.
- Keep locks out of network, database, filesystem, command, and notification
  I/O. Avoid global contention and per-event goroutines on scheduler hot paths.
- When database concurrency or volume is affected, minimize SQLite write
  transactions and their duration, batch writes when safe, validate critical
  indexes, and test realistic concurrent reads/writes and busy/lock behavior.
- For migrations on populated tables, assess table scans, lock duration,
  temporary disk use, write amplification, and deploy-time availability on all
  three databases. Prefer resumable or batched backfills when necessary.
- Add a regression test or benchmark for a material performance fix or risk.
  If measurement is impractical, explain why and use a bounded substitute.
- Do not merge or release a material unexplained regression. A deliberate
  tradeoff requires explicit user approval and must be documented.
