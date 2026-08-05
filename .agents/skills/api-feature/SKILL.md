---
name: api-feature
description: Implement or review an end-to-end gocron HTTP API change. Use when adding or changing Gin routes, handlers, request or response payloads, authorization, audit behavior, frontend API clients, API types, OpenAPI-style documentation, or API tests.
---

# Build a gocron API feature

Keep the route, authorization rule, handler, frontend client, types,
translations, documentation, and tests as one change.

## Compatibility invariants

- When the contract changes, preserve both deployment orders: an `N`
  frontend/client must tolerate an `N-1` backend, and an `N` backend must serve
  `N-1` frontends/clients.
- Keep routes, methods, accepted requests, response fields and types, status
  codes, pagination, and SSE event contracts stable. New response fields must
  be optional; new request fields must have compatible defaults.
- Detect new server capabilities before depending on them. If `N-1` lacks an
  endpoint, field, or stream behavior, retain a safe legacy path or hide it.
- Never repurpose an existing field or error code. Breaking contracts require
  explicit approval, a major version, and migration/rollout/rollback plans.

## Performance invariants

- Bound every collection response and expensive input. Require pagination,
  maximum page sizes, upload/body limits, and bounded filters where applicable.
- Prevent N+1 queries and per-row remote calls. Capture query counts for list
  endpoints and inspect query plans for new filters, sorting, or joins.
- For polling and SSE, bound connection/resource use, clean up on disconnect,
  avoid per-client database polling where sharing is safe, and apply
  backpressure or a documented drop/coalescing policy.
- Do not return or retain unbounded task output/logs. Use paging, cursors,
  bounded chunks, or streaming with cancellation.
- Benchmark materially affected high-traffic or high-volume endpoints with
  representative data and report applicable metrics.

## Trace the existing path

- Inspect neighboring registrations in `internal/routers/routers.go`, handlers
  under `internal/routers`, models/services called by the handler, and the
  matching frontend module under `web/gocronx-admin/src/api`.
- Identify every middleware applied to the route group. Classify the endpoint
  as public, authenticated, admin-only, API-token accessible, or agent-facing.
- Search path allowlists and permission maps before adding a route. Never make
  an endpoint public merely to make a request succeed.
- Preserve the response conventions in `internal/routers/base`. Do not expose
  raw database, filesystem, command, provider, or secret errors to clients.

## Implement the complete contract

- Use the appropriate HTTP method. Mutating operations must not use `GET`.
- Bind into an explicit request type, validate required fields and bounds, and
  reject unknown or dangerous input where the existing API pattern permits.
- Enforce authorization server-side before loading or mutating protected data.
  For resource ids, verify access to the resource rather than only validating
  that the caller is logged in.
- Use transactions for multi-write operations. Add an audit event for
  security-sensitive or administrative mutations following existing patterns.
- Update the frontend API wrapper and `src/types/api/api.d.ts` when the UI uses
  the endpoint. Keep backend and frontend field names/types aligned.
- Add both Chinese and English strings when user-visible text changes:
  backend `internal/modules/i18n/{zh_cn,en_us}.go`, frontend
  `src/locales/langs/{zh,en}.json`.
- Update `docs/zh/guide/api.md` for a public API contract. Include auth,
  parameters, example response, error behavior, and compatibility impact.

## Test risk, not just success

Add focused tests for:

- valid requests and stable response shape;
- missing, malformed, boundary, and nonexistent-resource input;
- unauthenticated and unauthorized callers;
- ownership or tenant boundary where applicable;
- duplicate submissions or retry behavior for mutations;
- secret/error redaction for sensitive endpoints.
- `N-1` requests/responses and both frontend/backend deployment orders when
  the feature crosses that boundary.

Run the changed router package with race detection, then relevant service/model
tests. Run frontend type checking when the contract is consumed by the UI.
Finally invoke `$verify` before committing.

Report the route and method, authorization class, contract changes, backward
compatibility, tests added, performance evidence when relevant, and any
documentation or i18n files intentionally left unchanged.
