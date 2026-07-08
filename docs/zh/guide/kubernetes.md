# Kubernetes 部署

gocron 提供 Helm Chart，支持一键部署到 Kubernetes 集群。

## 前置要求

- Kubernetes 1.19+
- Helm 3.0+

## 安装

### 添加 Helm 仓库

```bash
helm repo add gocron https://gocronx-team.github.io/gocron
helm repo update
```

### 快速安装

```bash
# 默认使用 SQLite，最简配置
helm install gocron gocron/gocron
```

### 使用自定义配置

```bash
# 方式一：命令行参数
helm install gocron gocron/gocron \
  --set db.engine=mysql \
  --set db.host=mysql.default \
  --set db.port=3306 \
  --set db.user=gocron \
  --set db.password=your_password \
  --set db.database=gocron

# 方式二：values 文件
helm install gocron gocron/gocron -f my-values.yaml
```

### 升级

```bash
helm upgrade gocron gocron/gocron --set image.tag=1.5.9
```

### 卸载

```bash
helm uninstall gocron
```

## 配置项

### 镜像

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `image.repository` | 镜像地址 | `gocronx/gocron` |
| `image.tag` | 镜像标签 | Chart appVersion |
| `image.pullPolicy` | 拉取策略 | `IfNotPresent` |

### 数据库

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `db.engine` | 数据库类型：`sqlite`、`mysql`、`postgres` | `sqlite` |
| `db.host` | 数据库地址 | `""` |
| `db.port` | 数据库端口 | `0` |
| `db.user` | 数据库用户 | `""` |
| `db.password` | 数据库密码 | `""` |
| `db.database` | 数据库名称/SQLite 文件路径 | `./data/gocron.db` |
| `db.prefix` | 表前缀 | `""` |
| `db.charset` | 字符集 | `utf8` |
| `db.maxIdleConns` | 最大空闲连接数 | `5` |
| `db.maxOpenConns` | 最大连接数 | `100` |

### 应用

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `app.name` | 应用名称 | `gocron` |
| `app.apiKey` | API Key | `""` |
| `app.apiSecret` | API Secret | `""` |
| `app.allowIps` | 允许访问的 IP | `""` |
| `app.concurrencyQueue` | 并发队列大小 | `500` |
| `app.enableTls` | 启用 TLS | `false` |
| `timezone` | 时区 | `Asia/Shanghai` |

### 服务

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `service.type` | 服务类型 | `ClusterIP` |
| `service.port` | 服务端口 | `5920` |

### Ingress

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `ingress.enabled` | 启用 Ingress | `false` |
| `ingress.className` | Ingress Class | `""` |
| `ingress.annotations` | 注解 | `{}` |
| `ingress.hosts` | 主机配置 | `[{host: gocron.local, paths: [{path: /, pathType: Prefix}]}]` |
| `ingress.tls` | TLS 配置 | `[]` |

### 持久化

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `persistence.enabled` | 启用持久化 | `true` |
| `persistence.storageClass` | 存储类 | `""` |
| `persistence.accessMode` | 访问模式 | `ReadWriteOnce` |
| `persistence.size` | 存储大小 | `1Gi` |

### 资源

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `resources.limits.cpu` | CPU 限制 | 不限制 |
| `resources.limits.memory` | 内存限制 | 不限制 |
| `resources.requests.cpu` | CPU 请求 | 不限制 |
| `resources.requests.memory` | 内存请求 | 不限制 |
| `replicaCount` | 副本数 | `1` |

::: warning 注意
使用 SQLite 时副本数必须为 1，因为 SQLite 不支持多进程并发写入。使用 MySQL 或 PostgreSQL 可设置多副本。
:::

## 部署示例

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

::: tip 内存限额与 GOMEMLIMIT
当你设置了 `resources.limits.memory`，gocron 启动时会自动把 Go 的 `GOMEMLIMIT`
设为该上限的约 90%（从容器 cgroup 读取）。这样垃圾回收会在 Pod 触顶前主动回收内存，
减少被 OOM kill，无需额外配置。未设内存限额时该行为为空操作（no-op）。
:::

### PostgreSQL + 多副本

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

## 注意事项

1. **SQLite 模式**：Deployment 策略自动设为 `Recreate`（而非 `RollingUpdate`），避免多 Pod 同时访问 SQLite 文件
2. **配置变更**：修改 ConfigMap 后 Pod 会自动重启（通过 checksum annotation 实现）
3. **数据持久化**：SQLite 模式下务必启用 PVC，否则 Pod 重启后数据丢失
4. **健康检查**：默认配置了 liveness 和 readiness 探针，通过 HTTP 检查 5920 端口
