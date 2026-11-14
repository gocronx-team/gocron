# Security hardening and compatible rollout

Existing configurations remain valid. Compatibility does **not** mean an
unauthenticated node or plaintext RPC channel is safe.

## Node RPC

For reverse-proxy deployments, set `GOCRON_PUBLIC_URL=https://scheduler.example`
on the scheduler to pin the origin used in generated installation commands and
scripts instead of trusting the incoming Host header. Only an HTTP(S) origin is
accepted (no credentials, path, query or fragment); invalid configuration rejects
installer generation. Leave it unset only when request-host discovery is intended.

- Never expose port 5921 publicly. Limit ingress to scheduler hosts.
- Configure mutual TLS on both scheduler and node using a dedicated CA and
  separate certificates. Keep certificate private keys readable only by the
  service account.
- Node `-require-secure -enable-tls -ca-file ... -cert-file ... -key-file ...`
  refuses startup without TLS; certificate loading and client verification
  remain mandatory. This flag is opt-in so older deployments still start.
- A shared `rpc_token` / `GOCRON_NODE_TOKEN` provides authentication, **not
  encryption**. Do not send tokens or task secrets over an untrusted network.
- The generated quick installer does not provision mutual TLS credentials.
  Review its service configuration before enabling network access.
  Do not embed shared tokens or certificate private keys in installer URLs or
  publicly served scripts. Provision credentials separately over a trusted
  channel; automated enrollment and certificate rotation are not implemented.
- Download release packages over trusted HTTPS and verify them against an
  independently trusted release digest/signature before installation. A digest
   fetched over the same compromised channel is not a trust anchor.

The quick installer accepts an optional `GOCRON_NODE_SHA256` from the installing
user's environment and verifies the downloaded archive **before extraction**.
Obtain that digest independently, review/download the script, then run
`GOCRON_NODE_SHA256=<trusted-64-hex-digest> bash install.sh`. Without the variable
the legacy flow remains available but prints an explicit integrity warning.
This does not authenticate the script itself: fetch it through trusted HTTPS.

Roll out certificates and scheduler configuration before requiring TLS on nodes.
Test one node first; keep a rollback copy of configuration and certificates.
Do not fall back to plaintext after a TLS or authentication failure.

## Outbound HTTP

`GOCRON_HTTP_PUBLIC_ONLY=1` on the scheduler process opts the shared HTTP task,
Webhook and Slack client into public-address-only connections. The check is
performed after DNS resolution and dials the checked numeric address, including
connections initiated by redirects. It blocks private, loopback, link-local,
multicast and carrier-grade NAT destinations. Existing connection pools require
a process restart when changing this environment variable.

Leave the variable unset for deployments that intentionally execute internal
HTTP jobs. In that mode, restrict who can edit tasks and notification URLs and
enforce an egress firewall/proxy allowlist, especially for cloud metadata.
This client policy is defense in depth, not a complete network sandbox: other
clients (including AI providers), routing, NAT and application-level relays need
separate controls. Do not mark all SSRF exposure resolved solely by enabling it.

## Temporary manual HTTP bodies

`GET /api/task/run/:id` is unchanged. New servers also support authenticated
`POST /api/task/run/:id` with `{"http_body":"{\"date\":\"2026-10-01\"}"}`.
Only HTTP POST tasks accept this override. The body must be valid JSON up to
65535 bytes; it is copied into this execution without saving task configuration.
Scheduling, concurrency and retry settings are unchanged. Audit records omit
the body, which may contain sensitive data.

New frontends first query `GET /api/task/run-capabilities`. If an older server
does not advertise `http_post_body_override`, only the original immediate run
is offered. A failed override POST must never be retried through the old GET
endpoint, because the task might already have been accepted.

## Minimal Prometheus endpoint

`GET /api/system/metrics` returns Prometheus text format through the existing
administrator authentication. Supply a valid administrator `Auth-Token` header
using Prometheus `http_headers` (or a secured local proxy that manages token
renewal). Do not expose an unauthenticated proxy. This initial endpoint exports
build information, process uptime and goroutine count only: no database queries,
per-task labels, historical execution counters or cross-HA aggregation. Scrape
each replica separately; business/task metrics remain a follow-up feature.
