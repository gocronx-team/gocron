# 后端国际化指南

## 使用方法

### 1. 在路由处理函数中使用

```go
import "github.com/gocronx-team/gocron/internal/modules/i18n"

func YourHandler(c *gin.Context) {
    // 使用国际化消息
    result := json.CommonFailure(i18n.T(c, "form_validation_failed"))
    c.String(http.StatusOK, result)
}
```

### 2. 添加新的翻译

在 `internal/modules/i18n/zh_cn.go` 添加中文：
```go
"your_key": "你的中文消息",
```

在 `internal/modules/i18n/en_us.go` 添加英文：
```go
"your_key": "Your English message",
```

### 3. 前端设置语言

前端需要在HTTP请求头中设置 `Accept-Language`:
- `zh-CN` 或 `zh` - 中文
- `en-US` 或 `en` - 英文

## 已完成的文件

- ✅ internal/routers/user/twofa.go

## 待迁移的文件

需要将以下文件中的硬编码中文消息替换为 i18n.T() 调用：

- internal/routers/install/install.go
- internal/routers/host/host.go
- internal/routers/manage/notification.go
- internal/routers/manage/system.go
- internal/routers/task/task.go
- internal/routers/user/user.go

## 常用消息键

- form_validation_failed - 表单验证失败
- user_not_found - 用户不存在
- update_success - 更新成功
- operation_success - 操作成功
- incomplete_parameters - 参数不完整
