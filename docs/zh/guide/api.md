# API 文档

gocron 提供了两类 API 接口：
1. **Open API**：`/api/v1/*`，使用签名认证。这是为后续第三方系统集成设计的接口规范（由于引入了 2FA，目前仅支持部分非敏感操作）。
2. **Web API**：`/api/*`，使用 Cookie/Session 认证，功能完整，适合深度定制或自动化脚本（需模拟登录）。

---

## 第一部分：Open API (签名认证)

这部分接口是为未来第三方集成预留的设计。由于系统引入了双因素认证 (2FA)，为了保证安全，目前 Open API 仅开放了部分非敏感操作接口。

### 认证机制

::: warning 前置条件
使用前请在 `conf/app.ini` 中配置 `api.key` 和 `api.secret`。
:::

**公共请求参数**：

| 参数名 | 类型 | 必填 | 说明 | 示例 |
|:---:|:---:|:---:|:---|:---|
| time | int64 | 是 | 当前 Unix 时间戳（秒） | `1495519926` |
| sign | string | 是 | 签名字符串（小写） | `418819b148815894e4abfa944d19f70f` |

**签名公式**：
```
sign = MD5(api.key + time + url_path + api.secret)
```

### 接口列表

#### 1. 启用任务
- **地址**: `/api/v1/task/enable/:id`
- **方式**: `POST`
- **说明**: 启用指定 ID 的任务

#### 2. 禁用任务
- **地址**: `/api/v1/task/disable/:id`
- **方式**: `POST`
- **说明**: 禁用指定 ID 的任务

#### 3. 删除任务日志
- **地址**: `/api/v1/tasklog/remove/:id`
- **方式**: `POST`
- **说明**: 删除指定 ID 的日志

---

## 第二部分：Web API (Cookie 认证)

这部分是管理后台使用的接口，功能最全。调用前需要先调用登录接口获取 Cookie。

### 认证机制

1. 调用登录接口 `/api/user/login`
2. 获取响应 Header 中的 `Set-Cookie`
3. 后续请求在 Header 中携带该 Cookie

### 1. 用户登录

- **地址**: `/api/user/login`
- **方式**: `POST`

**请求参数**：

| 参数名 | 类型 | 必填 | 说明 |
|:---:|:---:|:---:|:---|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

**返回示例**：
```json
{
    "code": 0,
    "message": "登录成功",
    "data": {
        "token": "..." // 这里的 token 主要用于前端状态，后端认证主要靠 Cookie
    }
}
```

### 2. 任务管理

#### 获取任务列表
- **地址**: `/api/task`
- **方式**: `GET`
- **参数**:
  - `page`: 页码 (默认 1)
  - `page_size`: 每页数量 (默认 20)
  - `id`: 任务ID (可选)
  - `name`: 任务名称 (可选)
  - `host_id`: 主机ID (可选)
  - `status`: 状态 (1:启用 0:停止)

#### 获取任务详情
- **地址**: `/api/task/:id`
- **方式**: `GET`

#### 新增/编辑任务
- **地址**: `/api/task/store`
- **方式**: `POST`
- **参数**:

| 参数名 | 类型 | 必填 | 说明 |
|:---:|:---:|:---:|:---|
| id | int | 否 | 任务ID (编辑时必填，新增为0) |
| name | string | 是 | 任务名称 |
| spec | string | 是 | Crontab 表达式 (秒级) |
| command | string | 是 | 执行命令 |
| protocol | int | 是 | 协议 (1:HTTP 2:Shell) |
| host_id | int | 是 | 主机ID |
| timeout | int | 否 | 超时时间(秒) |
| multi | int | 否 | 是否单实例运行 (1:是 2:否) |
| retry_times | int | 否 | 重试次数 |
| remark | string | 否 | 备注 |

#### 手动运行任务
- **地址**: `/api/task/run/:id`
- **方式**: `GET`

#### 删除任务
- **地址**: `/api/task/remove/:id`
- **方式**: `POST`

### 3. 主机管理

#### 获取主机列表
- **地址**: `/api/host`
- **方式**: `GET`
- **参数**:
  - `page`: 页码
  - `page_size`: 每页数量
  - `name`: 主机名称 (可选)

#### 新增/编辑主机
- **地址**: `/api/host/store`
- **方式**: `POST`
- **参数**:

| 参数名 | 类型 | 必填 | 说明 |
|:---:|:---:|:---:|:---|
| id | int | 否 | 主机ID (编辑时必填) |
| name | string | 是 | 主机名称 |
| alias | string | 是 | 别名 |
| port | int | 是 | 端口 (默认 5921) |
| remark | string | 否 | 备注 |

#### 删除主机
- **地址**: `/api/host/remove/:id`
- **方式**: `POST`

### 4. 日志管理

#### 获取任务日志
- **地址**: `/api/task/log`
- **方式**: `GET`
- **参数**:
  - `page`: 页码
  - `page_size`: 每页数量
  - `task_id`: 任务ID (可选)
  - `status`: 状态 (1:失败 2:执行中 3:成功 4:取消)

#### 停止任务执行
- **地址**: `/api/task/log/stop`
- **语义**: 成功表示所有目标节点均已确认收到停止信号。进程退出并由调度端完成收尾前，日志会短暂保持 `Running`，随后变为 `Cancelled`。如果所有节点均确认执行 ID 已不存在，残留的 `Running` 日志会自动修正为 `Cancelled`；节点不可用时返回错误且不修改数据库状态。
- **方式**: `POST`
- **参数**:
  - `id`: 日志ID
  - `task_id`: 任务ID

#### 实时读取任务输出
- **地址**: `/api/task/log/:id/stream`
- **方式**: `GET`
- **认证**: 与管理后台一致，携带登录令牌
- **参数**:
  - `id`: 日志ID
  - `seq`: 已接收的最后一个分片序号（可选，默认 0；断线重连时用于续读）
- **响应**: `text/event-stream`
  - `log`: `{"content":"...","seq":12,"status":1}`；`reset=true` 表示客户端应替换本地内容
  - `done`: 任务进入成功、失败或取消终态后发送，随后连接关闭
- **兼容性**: 新版执行节点流式上报；旧版节点自动回退为任务结束后一次性返回输出。
