# v1.13.0 release notes

## Features and fixes

- Allow a JSON body override for one manual HTTP POST task execution without
  persisting it. Capability detection preserves older backend behavior.
- Repair missing Webhook template records, propagate save errors and serialize
  concurrent creation across database-backed replicas with bounded retries.
- Persist the existing failure-diagnosis notification setting for tasks/templates.
- Add an administrator-only process-local Prometheus text endpoint.
- Make task-log query grouping explicit and add dependency-selection regressions.

## Compatible hardening

- Quote installer inputs, allow an explicitly configured public origin, and
  support an independently trusted optional node package digest.
- Add opt-in mutual-TLS enforcement for nodes and plaintext/authentication warnings.
- Add opt-in public-only outbound HTTP address validation with bounded connection
  fallback. Internal HTTP destinations remain supported by default.
- Update Go to 1.26.11 and refresh Go, frontend, documentation and commit tooling
  dependencies, retaining a tested CommonJS adapter for commit tooling.

See the bilingual security-hardening guides for deployment and remaining risks.
These changes do not provision node credentials or eliminate default plaintext,
unauthenticated RPC exposure. No schema or scheduler execution semantics changed.

## Upgrade notes

Back up the database and configuration before upgrading. This release introduces
no schema migration and preserves saved task bodies and scheduler semantics.
The Helm Chart is version 0.2.4 with application version 1.13.0.

The new manual body applies only to one HTTP POST execution, not to scheduled
runs, retries or saved task configuration. Older backends do not expose the new
input. Process metrics require administrator authentication and describe one
process, not an HA-wide or task-level aggregate.

Security options remain opt-in for compatibility: review the security-hardening
guide before exposing nodes or restricting outbound HTTP. Upgrade deployed
binaries and images explicitly; merging dependency updates does not patch
existing deployments.

## Verification

Go formatting, vet, lint and race tests; frontend build, types and lint; docs build;
dependency audits; and dedicated MySQL/PostgreSQL save tests have passed locally.
Real v1.11.0 mixed-version HA smoke tests and official v1.11.1 SQLite upgrade and
rollback retained tasks and saved bodies. Basic node unary and old MCP revision
contracts passed; this is not a complete TLS/stream/cancellation matrix.

The updated Dockerfile passed GitHub CI and a local isolated BuildKit build with
Go 1.26.11 using a trusted registry mirror. The default local Docker Hub route
still has connectivity problems; the mirror does not change deployment settings.
Release-candidate checks also include official checksum-verified v1.12.0 binaries:
SQLite, MySQL 8.4 and PostgreSQL 18.6 upgrade/rollback retained login, tasks and
saved JSON bodies. Mixed PostgreSQL HA smoke tests passed healthy leader handoff
in both directions, with successful short-window cron and parent/child execution.
Basic unary checks do not constitute a full
cross-version TLS, streaming or cancellation matrix.
