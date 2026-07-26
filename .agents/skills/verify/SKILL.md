---
name: verify
description: Reproduce the complete gocron CI pipeline locally and report every result. Use before commits, pull requests, merges, tags, or releases; when asked to verify, check, lint, test, build, or determine whether the repository is ready to ship.
---

# Verify gocron

Run from the repository root. Treat `.github/workflows/ci.yml` as the source of
truth if it differs from this skill.

## Preserve the working tree

- Inspect `git status --short` before running checks.
- Do not modify source files when the user only asks to verify or report.
- If the user also asks to fix failures, make the smallest relevant fixes and
  rerun every affected check.
- Do not discard, overwrite, or attribute pre-existing changes to this work.

## Run all checks

Do not stop after the first failure. Record the exit status and useful output
for each independent check.

1. Run `gofmt -l .`. Any listed Go file is a failure; do not run `gofmt -w`
   unless fixes were requested.
2. Run `go vet ./...`.
3. Run `golangci-lint run ./...`. CI uses v2.12.2; report a missing or
   incompatible local version instead of silently substituting another version.
4. Run `go test -race -coverprofile=coverage.out ./...`.
5. In `web/gocronx-admin`, run these as separate checks so one failure does not
   hide the others:
   - `pnpm build-only`
   - `pnpm exec vue-tsc --noEmit`
   - `pnpm lint`
6. Run `docker build -f Dockerfile.gocron -t gocron:ci .`.

If dependencies or tools are missing, clearly distinguish an environment
failure from a code failure. Installing tools or packages may require network
access; obtain any required approval and use the versions pinned by the repo or
CI. Do not claim the check passed when it was skipped.

## Report

End with a compact table containing each check, pass/fail/skipped status, and
the actionable error. Include exact `file:line` locations where available.
Mention whether `coverage.out` or generated frontend files changed the working
tree.

Say the repository is ready to commit, merge, or release only when every
required check passes, including the Docker build.
