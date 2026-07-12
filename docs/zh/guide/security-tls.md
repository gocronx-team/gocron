# 安全 - TLS 双向认证

TLS 双向认证为 gocron 提供了更高级别的安全保护，确保客户端和服务器之间的通信安全。

## 什么是 TLS 双向认证

TLS 双向认证（Mutual TLS Authentication）是一种安全机制，要求客户端和服务器都提供证书进行身份验证：
- **服务器认证**：客户端验证服务器的身份
- **客户端认证**：服务器验证客户端的身份

## 配置 TLS

### 1. 生成证书

首先需要生成 CA 证书、服务器证书和客户端证书。

**生成 CA 证书**：
```bash
# 生成 CA 私钥
openssl genrsa -out ca.key 2048

# 生成 CA 证书
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt
```

**生成服务器证书**：
```bash
# 生成服务器私钥
openssl genrsa -out server.key 2048

# 生成服务器证书签名请求
openssl req -new -key server.key -out server.csr

# 使用 CA 签名服务器证书
openssl x509 -req -days 3650 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
```

**生成客户端证书**：
```bash
# 生成客户端私钥
openssl genrsa -out client.key 2048

# 生成客户端证书签名请求
openssl req -new -key client.key -out client.csr

# 使用 CA 签名客户端证书
openssl x509 -req -days 3650 -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt
```

### 2. 配置 gocron

在配置文件 `.gocron/conf/app.ini` 中添加以下配置：

```ini
[tls]
enable_tls = true
ca_file = /path/to/ca.crt
cert_file = /path/to/server.crt
key_file = /path/to/server.key
```

### 3. 重启服务

修改配置后，重启 gocron 服务使配置生效。

## 客户端配置

使用 TLS 双向认证时，客户端也需要配置证书：

```bash
curl --cacert ca.crt --cert client.crt --key client.key https://gocron-server:5920
```

## 验证配置

可以使用以下命令验证 TLS 配置是否正确：

```bash
openssl s_client -connect localhost:5920 -CAfile ca.crt -cert client.crt -key client.key
```

## 节点 RPC 共享令牌（非 TLS 部署的鉴权）

TLS 双向认证需要签发和分发证书，运维成本较高。如果你暂时无法启用 mTLS，**强烈建议至少为节点 RPC 配置一个共享令牌**，否则在默认（无 TLS）配置下，任何能连到节点 `5921` 端口的人都可以在节点上执行任意命令并读取任务注入的机密。

共享令牌与 TLS **互相独立**，可以单独启用，也可以叠加在 TLS 之上作为额外一层校验。

### 工作原理

- **节点侧（验票方）**：配置令牌后，节点会对每一次收到的调用强制校验令牌（常量时间比较），不匹配直接拒绝。
- **调度器侧（发票方）**：配置令牌后，调度器会在每一次对节点的调用附带该令牌。
- 两端令牌值必须**完全一致**。

### 配置方式

**1. 调度器（gocron 主服务）** —— 在配置文件 `.gocron/conf/app.ini` 中设置：

```ini
rpc_token = 你的随机令牌
```

**2. 每个节点（gocron-node）** —— 通过命令行标志或环境变量设置（推荐用环境变量，避免令牌出现在进程列表中）：

```bash
# 方式一：命令行标志
./gocron-node -s 0.0.0.0:5921 -token 你的随机令牌

# 方式二：环境变量（推荐）
export GOCRON_NODE_TOKEN=你的随机令牌
./gocron-node -s 0.0.0.0:5921
```

令牌可以用任意足够随机的字符串，例如：`openssl rand -hex 32`。

### 升级与启用顺序（重要）

该功能默认关闭，**两端都不配置时行为与旧版本完全一致**，因此升级本身不会影响现有部署。但在**启用**令牌时必须遵循顺序：

1. **先**在调度器配置 `rpc_token` 并重启——此时调度器开始发令牌。旧节点会忽略它不认识的令牌，仍正常工作；无令牌的新节点也正常。
2. **再**逐台给节点配置 `-token` / `GOCRON_NODE_TOKEN` 并重启。

**切勿反过来**：如果节点先开启令牌校验，而调度器还没配置（不发令牌），该节点的任务会被全部拒绝（`Unauthenticated`）。

> 提示：调度器的令牌是全局的，对所有节点使用同一个值；不存在按节点分别配置令牌。

## 故障排除

### 常见问题

**Q: 启用 TLS 后无法访问**
- 检查证书路径是否正确
- 确认证书文件权限
- 查看 gocron 日志了解详细错误信息

**Q: 证书验证失败**
- 确认 CA 证书、服务器证书和客户端证书是否由同一个 CA 签发
- 检查证书是否过期
- 验证证书的 Common Name (CN) 是否正确

## 最佳实践

- **定期更新证书**：证书应定期更新，避免过期
- **安全存储私钥**：私钥文件应设置适当的权限，避免泄露
- **使用强加密**：使用 2048 位或更高的 RSA 密钥
- **监控证书有效期**：设置提醒，在证书过期前及时更新

## 相关文档

- [安全 - 双因素认证 (2FA)](./security-2fa)
- [配置文件](./configuration)
