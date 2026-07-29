#!/usr/bin/env bash

set -u

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"
pnpm_audit_store="${TMPDIR:-/tmp}/gocron-security-pnpm-store"
failures=0
skips=0

run_check() {
  local name="$1"
  shift
  printf '\n[%s]\n' "$name"
  if "$@"; then
    printf 'PASS: %s\n' "$name"
  else
    printf 'FAIL: %s\n' "$name"
    failures=$((failures + 1))
  fi
}

skip_check() {
  printf '\nSKIP: %s (%s)\n' "$1" "$2"
  skips=$((skips + 1))
}

cd "$repo_root" || exit 1

if command -v govulncheck >/dev/null 2>&1; then
  run_check "Go vulnerability scan" govulncheck ./...
else
  skip_check "Go vulnerability scan" "govulncheck is not installed"
fi

if command -v pnpm >/dev/null 2>&1; then
  run_check "root pnpm audit" env PNPM_CONFIG_STORE_DIR="$pnpm_audit_store" pnpm audit --audit-level high
  run_check "admin pnpm audit" env PNPM_CONFIG_STORE_DIR="$pnpm_audit_store" pnpm --dir web/gocronx-admin audit --audit-level high
  run_check "docs pnpm audit" env PNPM_CONFIG_STORE_DIR="$pnpm_audit_store" pnpm --dir docs audit --audit-level high
else
  skip_check "pnpm audits" "pnpm is not installed"
fi

if command -v gitleaks >/dev/null 2>&1; then
  if [[ "${GOCRON_SECURITY_SCAN_HISTORY:-0}" == "1" ]]; then
    run_check "secret scan (git history)" gitleaks detect --source . --no-banner --redact --verbose
  else
    run_check "secret scan (working tree)" gitleaks detect --source . --no-git --no-banner --redact --verbose
  fi
else
  skip_check "secret scan" "gitleaks is not installed"
fi

if command -v trivy >/dev/null 2>&1; then
  run_check "filesystem vulnerability scan" trivy fs --scanners vuln,secret --severity HIGH,CRITICAL --exit-code 1 .
else
  skip_check "filesystem vulnerability scan" "trivy is not installed"
fi

printf '\nSummary: %d failed, %d skipped\n' "$failures" "$skips"
if (( failures > 0 )); then
  exit 1
fi
if (( skips > 0 )); then
  exit 2
fi
