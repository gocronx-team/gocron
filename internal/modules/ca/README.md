# CA 证书管理机制

## 概述

Gocron 使用统一的 CA（证书颁发机构）来管理所有节点的 TLS 证书，实现安全的双向认证通信。

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                      Gocron Server                          │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              CA Certificate Authority                │   │
│  │  - ca.crt (CA 根证书)                                │   │
│  │  - ca.key (CA 私钥)                                  │   │
│  │  - client.crt (客户端证书，用于连接 agent)           │   │
│  │  - client.key (客户端私钥)                           │   │
│  └──────────────────────────────────────────────────────┘   │
│                          │                                   │
│                          │ 签发证书                          │
│                          ▼                                   │
└─────────────────────────────────────────────────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│  Agent Node  │   │  Agent Node  │   │  Agent Node  │
│  - ca.crt    │   │  - ca.crt    │   │  - ca.crt    │
│  - server.crt│   │  - server.crt│   │  - server.crt│
│  - server.key│   │  - server.key│   │  - server.key│
└──────────────┘   └──────────────┘   └──────────────┘
```

## 证书类型

### 1. CA 证书（ca.crt + ca.key）
- **位置**: `~/.gocron/ca/ca.crt` 和 `~/.gocron/ca/ca.key`
- **用途**: 作为根证书颁发机构，签发所有客户端和服务端证书
- **生命周期**: 10 年
- **生成时机**: Gocron 服务端首次启动时自动生成

### 2. 客户端证书（client.crt + client.key）
- **位置**: `~/.gocron/ca/client.crt` 和 `~/.gocron/ca/client.key`
- **用途**: Gocron 服务端连接 agent 节点时使用
- **生命周期**: 10 年
- **生成时机**: CA 初始化时自动生成
- **特点**: 所有 agent 节点共享同一个客户端证书

### 3. 服务端证书（server.crt + server.key）
- **位置**: Agent 节点的 `certs/` 目录
- **用途**: Agent 节点提供 gRPC 服务时使用
- **生命周期**: 10 年
- **生成时机**: Agent 节点注册时动态生成
- **特点**: 每个 agent 节点有独立的证书，包含节点的 IP 地址

## 工作流程

### 服务端启动流程

1. **初始化 CA**
   ```go
   // cmd/gocron/gocron.go
   func initCA() {
       _ = ca.GetGlobalCA()  // 触发 CA 初始化
   }
   ```

2. **加载或创建 CA**
   - 检查 `~/.gocron/ca/ca.crt` 是否存在
   - 如果存在，加载现有 CA
   - 如果不存在，创建新的 CA 并保存

3. **生成客户端证书**
   - 检查 `~/.gocron/ca/client.crt` 是否存在
   - 如果不存在，使用 CA 签发客户端证书

### Agent 注册流程

1. **Agent 发起注册请求**
   ```bash
   curl -X POST https://gocron-server/api/host/provision \
     -H "X-Register-Token: xxx" \
     -d '{"hostname":"node1","ip":"192.168.1.100"}'
   ```

2. **服务端生成证书**
   ```go
   // internal/routers/host/provision.go
   globalCA := ca.GetGlobalCA()
   serverCertPEM, serverKeyPEM, err := globalCA.GenerateServerCert(ip, hostname)
   ```

3. **返回证书包**
   ```json
   {
     "code": 0,
     "data": {
       "cert_bundle": {
         "ca_cert": "-----BEGIN CERTIFICATE-----...",
         "server_cert": "-----BEGIN CERTIFICATE-----...",
         "server_key": "-----BEGIN RSA PRIVATE KEY-----..."
       }
     }
   }
   ```

4. **Agent 保存证书**
   - 保存到 `certs/ca.crt`
   - 保存到 `certs/server.crt`
   - 保存到 `certs/server.key`

### 双向认证流程

1. **Gocron 连接 Agent**
   ```go
   // internal/modules/rpc/grpcpool/grpc_pool.go
   certificate := auth.Certificate{
       CAFile:     ca.GetCACertPath(),           // 用于验证 agent 证书
       CertFile:   ca.GetClientCertPath(),       // 客户端证书
       KeyFile:    ca.GetClientKeyPath(),        // 客户端私钥
       ServerName: agentIP,
   }
   ```

2. **Agent 验证客户端**
   - Agent 使用 `ca.crt` 验证 Gocron 的客户端证书
   - 验证通过后建立连接

3. **Gocron 验证服务端**
   - Gocron 使用 `ca.crt` 验证 Agent 的服务端证书
   - 验证通过后执行任务

## 安全特性

### 1. 双向认证
- Gocron 验证 Agent 的身份（通过服务端证书）
- Agent 验证 Gocron 的身份（通过客户端证书）
- 防止中间人攻击

### 2. 证书隔离
- 每个 Agent 节点有独立的服务端证书
- 证书包含节点的 IP 地址，防止证书被盗用

### 3. 统一管理
- 所有证书由同一个 CA 签发
- 便于证书的撤销和更新

## API 接口

### GetGlobalCA()
获取全局 CA 实例（单例模式）

```go
ca := ca.GetGlobalCA()
```

### GenerateServerCert(ip, hostname string)
为 Agent 节点生成服务端证书

```go
certPEM, keyPEM, err := ca.GenerateServerCert("192.168.1.100", "node1")
```

### GenerateClientCert()
生成客户端证书（用于 Gocron 连接 Agent）

```go
certPEM, keyPEM, err := ca.GenerateClientCert()
```

### GetCACertPath()
获取 CA 证书路径

```go
caPath := ca.GetCACertPath()
```

### GetClientCertPath()
获取客户端证书路径

```go
certPath, keyPath := ca.GetClientCertPath()
```

## 文件结构

```
~/.gocron/
└── ca/
    ├── ca.crt          # CA 根证书（公开）
    ├── ca.key          # CA 私钥（保密）
    ├── client.crt      # 客户端证书（Gocron 使用）
    └── client.key      # 客户端私钥（Gocron 使用）

/opt/gocron-node/
└── certs/
    ├── ca.crt          # CA 根证书（从服务端获取）
    ├── server.crt      # 服务端证书（从服务端获取）
    └── server.key      # 服务端私钥（从服务端获取）
```

## 证书更新

### 更新 CA 证书
1. 停止 Gocron 服务
2. 删除 `~/.gocron/ca/` 目录
3. 重启 Gocron 服务（自动生成新的 CA）
4. 重新注册所有 Agent 节点

### 更新 Agent 证书
1. 删除 Agent 节点的 `certs/` 目录
2. 重新运行安装脚本或手动调用 provision API

## 故障排查

### 问题：连接 Agent 失败，提示证书验证错误

**原因**: CA 证书不匹配

**解决方案**:
1. 检查 Gocron 的 `~/.gocron/ca/ca.crt`
2. 检查 Agent 的 `certs/ca.crt`
3. 确保两者内容一致
4. 如果不一致，重新注册 Agent

### 问题：Agent 启动失败，提示证书加载错误

**原因**: 证书文件损坏或不存在

**解决方案**:
1. 检查 `certs/` 目录下的文件是否完整
2. 重新运行安装脚本获取新证书

### 问题：证书过期

**原因**: 证书有效期为 10 年，到期后需要更新

**解决方案**:
1. 按照"更新 CA 证书"流程操作
2. 或者修改代码中的证书有效期

## 最佳实践

1. **定期备份 CA 证书**
   ```bash
   cp -r ~/.gocron/ca ~/backup/ca-$(date +%Y%m%d)
   ```

2. **保护 CA 私钥**
   - 确保 `ca.key` 文件权限为 600
   - 不要将 CA 私钥提交到版本控制系统

3. **使用环境变量**
   - 可以通过环境变量指定 CA 目录位置
   - 便于在容器环境中使用

4. **监控证书有效期**
   - 建议在证书到期前 1 年进行更新
   - 可以添加监控脚本检查证书有效期

## 测试

运行单元测试：

```bash
cd internal/modules/ca
go test -v
```

测试覆盖：
- CA 创建和加载
- 服务端证书生成
- 客户端证书生成
- 证书验证
- 单例模式
