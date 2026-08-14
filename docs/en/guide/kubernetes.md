# Kubernetes Deployment

The gocron Helm Chart deploys gocron as a managed, horizontally scalable
application. Kubernetes Service load-balances all Ready web/API Pods, while
database-backed leader election ensures that only one Pod runs the scheduler.

## Requirements

- Kubernetes 1.23+
- Helm 3
- An existing MySQL or PostgreSQL database and one stable connection endpoint

SQLite is intentionally not supported by the Kubernetes Chart. The standalone
binary continues to support SQLite and the web installation wizard.

## Install

```bash
helm repo add gocron https://gocronx-team.github.io/gocron
helm repo update
```

Use a values file so the database password is not exposed in shell history:

```yaml
# values-gocron.yaml
replicaCount: 2

db:
  engine: postgres
  host: postgresql.database.svc.cluster.local
  port: 5432
  user: gocron
  password: replace-me
  database: gocron
```

```bash
helm install gocron gocron/gocron -f values-gocron.yaml
```

The database must already exist. On first startup, the replicas coordinate
through a database advisory lock. One Pod creates the schema and administrator;
the others wait, then all start with the same configuration.

The Kubernetes deployment does not show the web installation wizard. The
administrator defaults to `admin`; its generated password is stored in the
release Secret:

```bash
kubectl get secret gocron -o jsonpath='{.data.admin-password}' | base64 -d; echo
```

## Existing Secret

Production deployments can provide a Secret instead of storing credentials in
Helm values:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gocron-runtime
type: Opaque
stringData:
  db-password: replace-me
  auth-secret: a-long-random-shared-jwt-secret
  encryption-key: a-long-random-encryption-key
  admin-password: replace-me
```

```yaml
db:
  engine: postgres
  host: postgresql.database.svc.cluster.local
  port: 5432
  user: gocron
  database: gocron

managed:
  existingSecret: gocron-runtime
  admin:
    username: admin
    email: admin@example.com
```

All replicas must use the same `auth-secret` and `encryption-key`. The bootstrap
administrator values are only used when the target database has no users.

## Scaling And Load Balancing

Scale manually from Helm, kubectl, or KubeVision:

```bash
kubectl scale deployment gocron --replicas=4
```

No configuration copy or shared PVC is required. Service/gocron automatically
distributes requests across Ready Pods. All Pods use the same database; only
the elected leader schedules tasks.

Optional CPU-based autoscaling:

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 6
  targetCPUUtilizationPercentage: 75
```

The Chart enables a PodDisruptionBudget and preferred cross-node pod
anti-affinity by default.

## Configuration

| Parameter | Description | Default |
|---|---|---|
| `replicaCount` | Initial number of Pods | `2` |
| `db.engine` | `mysql` or `postgres` | `postgres` |
| `db.host` | Stable database endpoint | required |
| `db.port` | Database port | `5432` |
| `db.user` | Database user | `gocron` |
| `db.password` | Database password for a Chart-managed Secret | required on first install |
| `db.database` | Existing database name | `gocron` |
| `managed.existingSecret` | Existing runtime Secret | `""` |
| `managed.admin.username` | Bootstrap administrator | `admin` |
| `managed.admin.password` | Bootstrap password; generated when empty | `""` |
| `managed.admin.email` | Bootstrap administrator email | `admin@example.com` |
| `autoscaling.enabled` | Create an HPA | `false` |
| `podDisruptionBudget.enabled` | Protect availability during eviction | `true` |
| `service.type` | Kubernetes Service type | `ClusterIP` |
| `ingress.enabled` | Create an Ingress | `false` |

## Upgrade

::: warning Chart 0.2.0
Chart 0.2.0 removes Kubernetes SQLite and the application PVC. Do not upgrade a
0.1.x SQLite release in place. Export its data to MySQL/PostgreSQL first, then
install or upgrade with the external database values. Back up the old PVC until
the migrated deployment has been verified.
:::

```bash
helm upgrade gocron gocron/gocron --reuse-values
```

Generated secrets are retained across upgrades. Config or Secret changes alter
the Deployment checksum and trigger a rolling update. Database schema changes
run once under the database bootstrap lock before each Pod joins the service.

## Uninstall

```bash
helm uninstall gocron
```

The external database is not deleted by Helm.
