# Security - TLS Mutual Authentication

TLS mutual authentication provides a higher level of security for gocron, ensuring secure communication between clients and servers.

## What is TLS Mutual Authentication

TLS Mutual Authentication is a security mechanism that requires both client and server to provide certificates for authentication:
- **Server Authentication**: Client verifies the server's identity
- **Client Authentication**: Server verifies the client's identity

## Configuring TLS

### 1. Generate Certificates

First, you need to generate CA certificate, server certificate, and client certificate.

**Generate CA Certificate**:
```bash
# Generate CA private key
openssl genrsa -out ca.key 2048

# Generate CA certificate
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt
```

**Generate Server Certificate**:
```bash
# Generate server private key
openssl genrsa -out server.key 2048

# Generate server certificate signing request
openssl req -new -key server.key -out server.csr

# Sign server certificate with CA
openssl x509 -req -days 3650 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
```

**Generate Client Certificate**:
```bash
# Generate client private key
openssl genrsa -out client.key 2048

# Generate client certificate signing request
openssl req -new -key client.key -out client.csr

# Sign client certificate with CA
openssl x509 -req -days 3650 -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt
```

### 2. Configure gocron

Add the following configuration to the configuration file `.gocron/conf/app.ini`:

```ini
[tls]
enable_tls = true
ca_file = /path/to/ca.crt
cert_file = /path/to/server.crt
key_file = /path/to/server.key
```

### 3. Restart Service

After modifying the configuration, restart the gocron service for the changes to take effect.

## Client Configuration

When using TLS mutual authentication, clients also need to configure certificates:

```bash
curl --cacert ca.crt --cert client.crt --key client.key https://gocron-server:5920
```

## Verify Configuration

You can use the following command to verify that the TLS configuration is correct:

```bash
openssl s_client -connect localhost:5920 -CAfile ca.crt -cert client.crt -key client.key
```

## Node RPC Shared Token (authentication for non-TLS deployments)

Mutual TLS requires issuing and distributing certificates, which has a higher operational cost. If you cannot enable mTLS for now, **it is strongly recommended to at least configure a shared token for node RPC**. Otherwise, in the default (no-TLS) configuration, anyone able to reach a node's `5921` port can execute arbitrary commands on the node and read the secrets injected into tasks.

The shared token is **independent of TLS**: it can be enabled on its own, or stacked on top of TLS as an extra verification layer.

### How it works

- **Node side (verifier)**: once a token is configured, the node enforces the token on every incoming call (constant-time comparison) and rejects any mismatch.
- **Scheduler side (issuer)**: once a token is configured, the scheduler attaches it to every call it makes to a node.
- The token values on both sides must be **exactly equal**.

### Configuration

**1. Scheduler (main gocron service)** — set it in the config file `.gocron/conf/app.ini`:

```ini
rpc_token = your-random-token
```

**2. Each node (gocron-node)** — set it via a command-line flag or environment variable (the env var is recommended so the token does not appear in the process list):

```bash
# Option 1: command-line flag
./gocron-node -s 0.0.0.0:5921 -token your-random-token

# Option 2: environment variable (recommended)
export GOCRON_NODE_TOKEN=your-random-token
./gocron-node -s 0.0.0.0:5921
```

Use any sufficiently random string as the token, e.g. `openssl rand -hex 32`.

### Upgrade & enablement order (important)

This feature is off by default. **When neither side configures a token, behavior is identical to older versions**, so upgrading itself does not affect existing deployments. But when **enabling** the token you must follow this order:

1. **First** configure `rpc_token` on the scheduler and restart it — now the scheduler starts sending the token. Old nodes ignore the unknown token and keep working; token-less new nodes also keep working.
2. **Then** configure `-token` / `GOCRON_NODE_TOKEN` on each node one by one and restart them.

**Do not do it the other way around**: if a node starts verifying tokens while the scheduler is not yet configured (and sends none), all tasks to that node will be rejected (`Unauthenticated`).

> Note: the scheduler token is global — the same value is used for all nodes. There is no per-node token configuration.

## Troubleshooting

### Common Issues

**Q: Cannot access after enabling TLS**
- Check if certificate paths are correct
- Confirm certificate file permissions
- Check gocron logs for detailed error information

**Q: Certificate verification failed**
- Confirm that CA certificate, server certificate, and client certificate are issued by the same CA
- Check if certificates are expired
- Verify that the Common Name (CN) in certificates is correct

## Best Practices

- **Regular certificate updates**: Certificates should be updated regularly to avoid expiration
- **Secure private key storage**: Private key files should have appropriate permissions to prevent leakage
- **Use strong encryption**: Use 2048-bit or higher RSA keys
- **Monitor certificate validity**: Set reminders to update certificates before expiration

## Related Documentation

- [Security - Two-Factor Authentication (2FA)](./security-2fa)
- [Configuration](./configuration)
