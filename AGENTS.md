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

## Compatibility policy (MANDATORY)

Backward compatibility is the default and is not optional.

- Release `N` must interoperate with the previous minor `N-1` wherever a change
  crosses frontend/backend, gocron/gocron-node, schema, protocol, or HA boundaries.
- Existing installations must upgrade in place without data loss on SQLite,
  MySQL, and PostgreSQL; never require a clean install.
- Keep HTTP/SSE contracts and scheduler semantics stable. Evolve protobuf/gRPC
  additively and use capability detection plus fallback for new behavior.
- Test affected boundaries with previous-version fixtures, contracts, or protocol
  doubles. Current-only and fresh-install-only tests are insufficient evidence.
- A breaking change requires explicit user approval before implementation, a
  major version, and migration, rollout, and rollback plans. Never hide one in
  a minor or patch release. If compatibility is uncertain, stop and ask.

## Performance policy (MANDATORY)

Material performance risk requires evidence, not intuition.

- For scheduler hot paths, high-volume APIs/logs/streams, large queries, or
  populated-table migrations, define a representative workload and compare
  applicable before/after metrics. Report the command, data size, and result.
- Do not introduce N+1 queries, unbounded reads, responses, logs, goroutines,
  queues, retries, polling, caches, or in-memory accumulation. Require limits,
  pagination/batching, cancellation, cleanup, and backpressure where relevant.
- Keep locks out of network, database, filesystem, command, and notification
  I/O. Avoid global contention and per-event goroutines on scheduler hot paths.
- Test SQLite write frequency, transaction duration, indexes, and lock behavior
  when database concurrency or volume is affected. Assess large migrations on
  all three databases and batch or resume backfills when necessary.
- Add a regression test or benchmark for a material performance fix or risk.
  Do not merge an unexplained material regression; approved tradeoffs must be
  documented.

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

## Vibe Coding & Agent Autonomy Guidelines

### 1. Codebase Feature Map
- **New Notification Channel**:
  - Backend sender: `internal/modules/notify/`
  - Backend Task model field: `NotifyType` in `internal/models/task.go`
  - Backend & Frontend i18n: `internal/modules/i18n/{zh_cn,en_us}.go` & `web/gocronx-admin/src/locales/langs/{zh,en}.json`
  - Frontend View & API: `web/gocronx-admin/src/views/system/notification/` & `web/gocronx-admin/src/api/notification.ts`
- **Database Schema Change**:
  - Model: `internal/models/<model>.go`
  - Migration (MANDATORY): `internal/models/migration.go` (Add model to `Install` tables + append new `versionId` & `upgradeForNNN` + add test).
- **New HTTP API Route**:
  - Backend Router: `internal/routers/<domain>/`
  - Backend Logic: `internal/service/`
  - Frontend Client: `web/gocronx-admin/src/api/<domain>.ts`

### 2. Autonomous Decision Boundaries
- **Auto-Approve (Proceed without prompting)**:
  - Adding/updating unit & integration tests.
  - Internal refactoring, fixing linter/go vet issues, code formatting.
  - Adding missing i18n key-values (must keep zh/en in sync).
  - Minor UI styling, Element Plus component wrapping, TailwindCSS adjustments.
- **Must Ask User First**:
  - Breaking HTTP/gRPC contracts, removing response fields, changing HTTP status codes.
  - Modifying scheduler execution semantics, lock mechanisms, or status transitions.
  - Database schema drops/renames or breaking previous minor version (N-1) compatibility.
  - Adding new heavy external infrastructure dependencies (e.g., Redis, Kafka).

### 3. Self-Healing & Clean Commit Rules
- **Fix Order**: Format (`gofmt`) -> Lint (`golangci-lint` / `pnpm lint`) -> Type Check (`vue-tsc`) -> Tests (`go test -race`).
- **Hygiene**: Remove temporary scratch scripts, debug `fmt.Println`s, and clean git untracked artifacts before requesting review or committing.

