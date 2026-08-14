# Kubernetes 部署

gocron Helm Chart 会将 gocron 部署为托管的、可横向扩容的应用。Kubernetes
Service 自动在所有 Ready 的 Web/API Pod 之间分发流量；基于数据库的 Leader
选举保证任何时刻只有一个 Pod 运行调度器。

## 前置要求

- Kubernetes 1.23+
- Helm 3
- 已创建的 MySQL 或 PostgreSQL 数据库，以及一个稳定的数据库访问地址

Kubernetes Chart 不支持 SQLite。独立二进制仍然支持 SQLite 和 Web 安装向导。

## 安装

```bash
helm repo add gocron https://gocronx-team.github.io/gocron
helm repo update
```

建议使用 values 文件，避免数据库密码出现在 Shell 历史记录中：

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

managed:
  authSecret: replace-with-at-least-32-random-characters
  encryptionKey: replace-with-another-32-random-characters
  admin:
    username: admin
    password: replace-with-a-strong-admin-password
    email: admin@example.com
```

```bash
helm install gocron gocron/gocron -f values-gocron.yaml
```

数据库必须提前创建。首次启动时，各副本通过数据库 advisory lock 协调；一个 Pod
创建表结构和管理员，其他 Pod 等待初始化完成，然后使用同一份配置启动。

Kubernetes 部署不会显示 Web 安装向导。当 Chart 管理运行时 Secret 时，必须显式填写
上面的四项敏感配置，使模板渲染结果稳定，并兼容 KubeVision 这类带预览确认的控制台。
配置的管理员密码保存在 Release Secret 中：

```bash
kubectl get secret gocron -o jsonpath='{.data.admin-password}' | base64 -d; echo
```

## 使用已有 Secret

生产环境可以预先创建 Secret，避免把凭据写入 Helm values：

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

所有副本必须使用相同的 `auth-secret` 和 `encryption-key`。只有目标数据库中不存在用户
时，启动过程才会使用管理员初始化配置。

## 扩容与负载均衡

可以通过 Helm、kubectl 或 KubeVision 扩容：

```bash
kubectl scale deployment gocron --replicas=4
```

扩容不需要复制配置，也不需要共享 PVC。Service/gocron 会自动将请求分发给 Ready
Pod；所有 Pod 使用同一个数据库，只有选举出的 Leader 负责调度任务。

可以启用基于 CPU 的自动扩容：

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 6
  targetCPUUtilizationPercentage: 75
```

Chart 默认启用 PodDisruptionBudget，并配置优先跨节点分散 Pod 的反亲和性。

## 配置项

| 参数                          | 说明                                 | 默认值                         |
| ----------------------------- | ------------------------------------ | ------------------------------ |
| `replicaCount`                | 初始 Pod 数量                        | `2`                            |
| `db.engine`                   | `mysql` 或 `postgres`                | `postgres`                     |
| `db.host`                     | 稳定的数据库访问地址                 | 必填                           |
| `db.port`                     | 数据库端口                           | `5432`                         |
| `db.user`                     | 数据库用户                           | `gocron`                       |
| `db.password`                 | Chart 管理 Secret 时使用的数据库密码 | 必填                           |
| `db.database`                 | 已存在的数据库名称                   | `gocron`                       |
| `managed.existingSecret`      | 已存在的运行时 Secret                | `""`                           |
| `managed.authSecret`          | 共享 JWT 密钥，至少 32 个字符        | 未配置 `existingSecret` 时必填 |
| `managed.encryptionKey`       | 共享加密密钥，至少 32 个字符         | 未配置 `existingSecret` 时必填 |
| `managed.admin.username`      | 初始化管理员                         | `admin`                        |
| `managed.admin.password`      | 初始化密码，至少 6 个字符            | 未配置 `existingSecret` 时必填 |
| `managed.admin.email`         | 初始化管理员邮箱                     | `admin@example.com`            |
| `autoscaling.enabled`         | 创建 HPA                             | `false`                        |
| `podDisruptionBudget.enabled` | 在驱逐期间保护可用性                 | `true`                         |
| `service.type`                | Kubernetes Service 类型              | `ClusterIP`                    |
| `ingress.enabled`             | 创建 Ingress                         | `false`                        |

## 升级

::: warning Chart 0.2.0
Chart 0.2.0 移除了 Kubernetes SQLite 和应用 PVC。不要直接升级使用 SQLite 的
0.1.x Release；应先将数据导入 MySQL/PostgreSQL，再使用外部数据库参数安装或升级。
在新部署验证完成前保留旧 PVC 备份。
:::

```bash
helm upgrade gocron gocron/gocron --reuse-values
```

升级时使用 `--reuse-values` 并保持运行时密钥不变。ConfigMap 或 Secret 发生变化时，
Deployment checksum 会触发滚动更新。数据库结构升级会在数据库初始化锁内执行一次，
完成后 Pod 才加入服务。

## 卸载

```bash
helm uninstall gocron
```

Helm 不会删除外部数据库。
