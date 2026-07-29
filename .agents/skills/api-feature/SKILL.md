---
name: api-feature
description: Implement or review an end-to-end gocron HTTP API change. Use when adding or changing Gin routes, handlers, request or response payloads, authorization, audit behavior, frontend API clients, API types, OpenAPI-style documentation, or API tests.
---

# Build a gocron API feature

Keep the route, authorization rule, handler, frontend client, types,
translations, documentation, and tests as one change.

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

Run the changed router package with race detection, then relevant service/model
tests. Run frontend type checking when the contract is consumed by the UI.
Finally invoke `$verify` before committing.

Report the route and method, authorization class, contract changes, backward
compatibility, tests added, and any documentation or i18n files intentionally
left unchanged.
