---
name: security-check
description: Audit or harden gocron security across Go, pnpm workspaces, containers, authentication, authorization, secrets, command execution, SSRF, and dependency vulnerabilities. Use for security reviews, vulnerability remediation, Dependabot security alerts, release hardening, or suspected exposure.
---

# Check gocron security

Default to read-only review. A request to scan or report does not authorize
dependency upgrades, source changes, alert dismissal, PR merges, or pushes.

## Establish scope and trust boundaries

Inspect the diff when reviewing a change; inspect the relevant data flow when
reviewing the repository. Prioritize internet-facing routes, API and agent
tokens, password/2FA flows, command execution, host/URL inputs, uploaded/imported
data, secret storage, logs, webhooks, AI provider calls, and RPC boundaries.

Check for:

- missing authentication, authorization, ownership checks, or audit events;
- SQL/command/template/path injection and unsafe shell construction;
- SSRF, unrestricted redirects, unsafe downloads, and weak URL validation;
- plaintext secrets, accidental logging, overly broad tokens, weak signing,
  insecure randomness, and missing expiry/rotation;
- mass assignment, unbounded input/body/queue sizes, brute force, and DoS;
- unsafe CORS/cookies/headers and frontend token exposure;
- vulnerable direct and transitive dependencies and unsafe container defaults.

Trace sanitizers and middleware to their implementation; do not infer safety
from function names. Do not print secret values while investigating.

## Run deterministic gates

Run from the repository root:

```bash
bash .agents/skills/security-check/scripts/security_gate.sh
```

The script runs independent checks and continues after failures. Missing tools
or network access are `SKIP`, never `PASS`. Review git changes afterward because
security tools must not silently alter lockfiles.

The default secret scan checks the current working tree. For the slower
full-history scan, run:

```bash
GOCRON_SECURITY_SCAN_HISTORY=1 bash .agents/skills/security-check/scripts/security_gate.sh
```

For authorization or input-validation changes, add focused negative tests and
run the affected package with `-race`. Invoke `$verify` after fixes.

## Triage and report

For each finding, provide severity, reachable attack path, affected
`file:line`, evidence, impact, and smallest safe remediation. Distinguish:

- confirmed exploitable behavior;
- defense-in-depth improvement;
- dependency advisory not reachable in this application;
- false positive with concrete justification.

Never dismiss or ignore an alert solely because tests pass. Do not claim the
repository is secure; state the scope covered and skipped checks.
