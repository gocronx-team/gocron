# 高可用部署

gocron 支持多实例部署，内置基于数据库行锁的 Leader 自动选举。只有 Leader 节点运行调度器，其他节点热备待命，Leader 故障后秒级自动接管。

## 工作原理

gocron 使用数据库行锁（`SELECT ... FOR UPDATE`）实现选主，参考了 [XXL-JOB](https://github.com/xuxueli/xxl-job) 在生产环境验证过的方案。

- 首次启动时自动创建 `scheduler_lock` 表
- 一个实例获取锁成为 **Leader**，运行所有定时任务
- 其他实例是 **Follower** — 提供 Web 界面和 API 服务，但不执行任务
- Leader 每 **5 秒**续约一次（租约 **15 秒**过期）
- Leader 优雅关机时立即释放锁 — 故障转移瞬间完成
- Leader 异常崩溃时，租约过期后 Follower 在 **15 秒**内自动接管

::: warning 数据库要求
高可用模式仅支持 **MySQL** 和 **PostgreSQL**。SQLite 不支持多进程并发访问，SQLite 模式下自动跳过选举（单节点运行）。
:::

## 部署步骤

### 1. 安装第一个节点

启动第一个节点，通过 Web 安装向导完成初始化。数据库请选择 **MySQL** 或 **PostgreSQL**。

```bash
./gocron web --port 5920
# 打开 http://localhost:5920 完成安装
```

### 2. 复制配置到其他节点

将第一个节点的 `.gocron/conf/` 目录复制到其他节点：

```bash
scp -r .gocron/conf/ user@node2:/path/to/gocron/.gocron/conf/
```

该目录包含：
- `app.ini` — 数据库和应用配置
- `install.lock` — 标记安装完成
- `.version` — 当前应用版本

### 3. 启动所有节点

```bash
# 节点 1
./gocron web --port 5920

# 节点 2
./gocron web --port 5921
```

两个节点连接同一数据库，其中一个会自动成为 Leader。

### 验证

查看日志：

```
# Leader 节点
This node elected as leader: node1:12345
Starting to load scheduled tasks (this node is leader)

# Follower 节点
Scheduler infrastructure initialized
# （没有 "elected as leader" 消息）
```

## 故障转移

### 优雅关机（Ctrl+C / SIGTERM）

```
Releasing leader lock: node1:12345
Stopping scheduler (this node lost leadership)
```

Follower **1 秒内**接管。

### 异常崩溃 / Kill -9

租约 15 秒后过期，Follower 检测到过期锁后自动接管。

## Kubernetes 部署

Kubernetes 环境使用官方 Helm Chart。Chart 会启用托管模式，将同一个认证 Secret 注入
所有 Pod，通过数据库 advisory lock 协调表结构初始化，并使用 Service 暴露所有 Ready
副本，不需要共享 PVC 或手动复制配置。

```bash
helm install gocron gocron/gocron \
  --set db.engine=postgres \
  --set db.host=postgresql.database.svc \
  --set db.port=5432 \
  --set db.user=gocron \
  --set db.password=replace-me \
  --set db.database=gocron
```

已有 Secret、扩容、HPA、Ingress 和管理员初始化配置请参考
[Kubernetes 部署](./kubernetes)。

## 架构图

```
┌─────────────┐     ┌─────────────┐
│   节点 1     │     │   节点 2     │
│  (Leader)   │     │ (Follower)  │
│             │     │             │
│ ┌─────────┐ │     │ ┌─────────┐ │
│ │ 调度器   │ │     │ │  待命    │ │
│ │ (运行中) │ │     │ │  (空闲)  │ │
│ └─────────┘ │     │ └─────────┘ │
│ ┌─────────┐ │     │ ┌─────────┐ │
│ │ Web 界面 │ │     │ │ Web 界面 │ │
│ │  & API  │ │     │ │  & API  │ │
│ └─────────┘ │     │ └─────────┘ │
└──────┬──────┘     └──────┬──────┘
       │                   │
       └───────┬───────────┘
               │
        ┌──────┴──────┐
        │  MySQL / PG │
        │             │
        │ scheduler   │
        │ _lock 表    │
        └─────────────┘
```
