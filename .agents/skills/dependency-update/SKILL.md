---
name: dependency-update
description: Review, apply, verify, or merge gocron dependency updates from Dependabot or manual requests. Use for Go modules, root/admin/docs pnpm packages, GitHub Actions, Docker base images, security advisories, lockfile conflicts, or dependency update pull requests.
---

# Update gocron dependencies

Handle one dependency group and ecosystem at a time unless a security fix
requires coordinated versions. A request to inspect a PR or alert is read-only;
merge, dismiss, commit, and push only when explicitly requested.

## Assess before changing

- Inspect the manifest and lockfile diff, release notes, advisory, dependency
  graph, and current CI status. Use primary upstream documentation.
- Classify the update as patch, minor, major, security, development-only, build
  tooling, GitHub Action, or runtime/container.
- Determine whether the dependency is direct, transitive, bundled into the
  frontend, shipped in the image, or used only in tests.
- For security alerts, record the vulnerable range, fixed version, severity,
  reachable code path, and whether every affected workspace is updated.
- Treat major updates, Go/toolchain changes, framework upgrades, database
  drivers, schedulers, RPC libraries, auth/crypto packages, and container bases
  as high risk even when the diff is small.

Never edit a lockfile by hand. Use the repository package manager and preserve
its existing lockfile format. Do not silence peer dependency failures or add
blanket advisory ignores to make a check green.

## Apply the smallest update

Use the narrowest ecosystem command that updates the requested package and its
required peers. Review the resulting diff for unrelated upgrades, install
scripts, new transitive packages, license changes, and unexpected engine or Go
version changes. Revert unrelated generated churn without discarding
pre-existing user changes.

For a Dependabot PR, prefer checking out or updating its branch rather than
reconstructing a different lockfile locally. If the branch is stale, rebase or
regenerate only when the user authorized changes to that PR.

## Verify by ecosystem

- Go runtime dependency: run `go mod tidy`, ensure `go.mod` and `go.sum` contain
  only expected changes, then run affected package tests with `-race`,
  `go vet ./...`, and `govulncheck ./...` when available.
- Admin frontend: run `pnpm --dir web/gocronx-admin build-only`,
  `pnpm --dir web/gocronx-admin exec vue-tsc --noEmit`, and
  `pnpm --dir web/gocronx-admin lint`.
- Docs: run `pnpm --dir docs docs:build`.
- Root tooling: run the affected commit/lint hook command and verify the admin
  workspace lockfile was not unintentionally changed.
- GitHub Action: validate YAML and inspect the referenced action's changelog and
  immutable tag/SHA policy.
- Docker/base image: build `Dockerfile.gocron` and inspect architecture,
  runtime user, binary/library compatibility, and vulnerability scan results.

Run the relevant package tests even if the dependency is marked dev-only.
Invoke `$verify` before merge for every code or lockfile update.

## Decide and report

State the exact old/new versions, why the update is needed, runtime exposure,
breaking-change review, manifest/lockfiles changed, audit result, checks run,
and remaining risk. Recommend merge only when required checks are green and the
diff contains no unexplained dependency churn.

If committing, use one single-line Conventional Commit subject under 100
characters and no `Co-Authored-By`, for example:

```text
chore(deps): update vue to 3.6.0
```
