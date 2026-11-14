# Iteration changes (unreleased)

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

## Verification and remaining gates

Go formatting, vet, lint and race tests; frontend build, types and lint; docs build;
dependency audits; and dedicated MySQL/PostgreSQL save tests have passed locally.
Real v1.11.0 mixed-version HA smoke tests and official v1.11.1 SQLite upgrade and
rollback retained tasks and saved bodies. Basic node unary and old MCP revision
contracts passed; this is not a complete TLS/stream/cancellation matrix.

The standard Docker build remains blocked by Docker Hub metadata connectivity in
the local environment. Do not treat the previous-toolchain image as validation
of this update. Complete Docker/CI gates before merge or release. This iteration
does not bump the application version, publish artifacts or close unresolved issues.
