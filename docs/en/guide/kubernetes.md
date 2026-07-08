# Kubernetes Deployment

gocron provides a Helm Chart for one-click deployment to Kubernetes clusters.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+

## Installation

### Add Helm Repository

```bash
helm repo add gocron https://gocronx-team.github.io/gocron
helm repo update
```

### Quick Install

```bash
# Uses SQLite by default, minimal configuration
helm install gocron gocron/gocron
```

### Custom Configuration

```bash
# Method 1: Command line parameters
helm install gocron gocron/gocron \
  --set db.engine=mysql \
  --set db.host=mysql.default \
  --set db.port=3306 \
  --set db.user=gocron \
  --set db.password=your_password \
  --set db.database=gocron

# Method 2: Values file
helm install gocron gocron/gocron -f my-values.yaml
```

### Upgrade

```bash
helm upgrade gocron gocron/gocron --set image.tag=1.5.9
```

### Uninstall

```bash
helm uninstall gocron
```

## Configuration

### Image

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Image repository | `gocronx/gocron` |
| `image.tag` | Image tag | Chart appVersion |
| `image.pullPolicy` | Pull policy | `IfNotPresent` |

### Database

| Parameter | Description | Default |
|-----------|-------------|---------|
| `db.engine` | Database type: `sqlite`, `mysql`, `postgres` | `sqlite` |
| `db.host` | Database host | `""` |
| `db.port` | Database port | `0` |
| `db.user` | Database user | `""` |
| `db.password` | Database password | `""` |
| `db.database` | Database name / SQLite file path | `./data/gocron.db` |
| `db.prefix` | Table prefix | `""` |
| `db.charset` | Character set | `utf8` |
| `db.maxIdleConns` | Max idle connections | `5` |
| `db.maxOpenConns` | Max open connections | `100` |

### Application

| Parameter | Description | Default |
|-----------|-------------|---------|
| `app.name` | Application name | `gocron` |
| `app.apiKey` | API Key | `""` |
| `app.apiSecret` | API Secret | `""` |
| `app.allowIps` | Allowed IPs | `""` |
| `app.concurrencyQueue` | Concurrency queue size | `500` |
| `app.enableTls` | Enable TLS | `false` |
| `timezone` | Timezone | `Asia/Shanghai` |

### Service

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `5920` |

### Ingress

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable Ingress | `false` |
| `ingress.className` | Ingress Class | `""` |
| `ingress.annotations` | Annotations | `{}` |
| `ingress.hosts` | Host configuration | `[{host: gocron.local, paths: [{path: /, pathType: Prefix}]}]` |
| `ingress.tls` | TLS configuration | `[]` |

### Persistence

| Parameter | Description | Default |
|-----------|-------------|---------|
| `persistence.enabled` | Enable persistence | `true` |
| `persistence.storageClass` | Storage class | `""` |
| `persistence.accessMode` | Access mode | `ReadWriteOnce` |
| `persistence.size` | Storage size | `1Gi` |

### Resources

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | none |
| `resources.limits.memory` | Memory limit | none |
| `resources.requests.cpu` | CPU request | none |
| `resources.requests.memory` | Memory request | none |
| `replicaCount` | Number of replicas | `1` |

::: warning Note
Replica count must be 1 when using SQLite, as SQLite does not support concurrent writes from multiple processes. Use MySQL or PostgreSQL for multiple replicas.
:::

## Deployment Examples

### SQLite + NodePort

```yaml
# values-sqlite.yaml
db:
  engine: sqlite

service:
  type: NodePort

persistence:
  enabled: true
  size: 2Gi
```

### MySQL + Ingress

```yaml
# values-mysql.yaml
db:
  engine: mysql
  host: mysql.database.svc
  port: 3306
  user: gocron
  password: your_password
  database: gocron

persistence:
  enabled: false

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: gocron.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: gocron-tls
      hosts:
        - gocron.example.com

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 256Mi
```

::: tip Memory limit & GOMEMLIMIT
When you set a `resources.limits.memory`, gocron automatically sets Go's
`GOMEMLIMIT` to about 90% of that limit at startup (read from the container's
cgroup). This makes the garbage collector reclaim memory before the pod hits its
limit, reducing OOM kills — no extra configuration needed. Without a memory
limit, this is a no-op.
:::

### PostgreSQL + Multiple Replicas

```yaml
# values-postgres.yaml
replicaCount: 3

db:
  engine: postgres
  host: pg.database.svc
  port: 5432
  user: gocron
  password: your_password
  database: gocron

persistence:
  enabled: false
```

## Notes

1. **SQLite mode**: Deployment strategy is automatically set to `Recreate` (instead of `RollingUpdate`) to prevent multiple Pods from accessing the SQLite file simultaneously
2. **Config changes**: Pods automatically restart when ConfigMap changes (via checksum annotation)
3. **Data persistence**: Always enable PVC in SQLite mode, otherwise data is lost when Pods restart
4. **Health checks**: Liveness and readiness probes are configured by default, checking HTTP on port 5920
