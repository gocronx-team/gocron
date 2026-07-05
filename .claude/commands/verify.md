---
description: Run the full CI pipeline locally (gofmt, vet, golangci-lint, go test -race, frontend build/typecheck/lint) and report what passes/fails. Use before committing or releasing.
---

Reproduce the CI pipeline (`.github/workflows/ci.yml`) locally so we never hit
"passed locally but failed CI". Run each step, and **do not stop at the first
failure** — collect all results and report a summary table at the end.

Run from the repo root:

1. **gofmt** — `gofmt -l .` (any output = unformatted files → fail; list them)
2. **go vet** — `go vet ./...`
3. **golangci-lint** — `golangci-lint run ./...` (v2.12.2, same as CI; this is a
   separate gate from `go test` — staticcheck/etc. issues fail here even when
   tests pass)
4. **go test (race)** — `go test -race ./...` (CI uses `-race`)
5. **frontend** — in `web/gocronx-admin`:
   `pnpm build-only && pnpm exec vue-tsc --noEmit && pnpm lint`

Notes:
- If `golangci-lint` isn't installed, say so and install with
  `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2`.
- Fix any failure with the smallest change that satisfies the linter/test, then
  re-run only the affected step.
- End with a table: step → ✅/❌, and for failures the exact file:line + message.
- Only report "ready to commit/release" when **every** step is green.
