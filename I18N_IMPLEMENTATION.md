# Gocron 国际化实现完成

## 概述

本项目已完成全面的国际化(i18n)支持，包括前端Vue应用和后端Go服务。

## 前端国际化 (Vue3 + vue-i18n)

### 已实现功能

1. **语言支持**
   - 简体中文 (zh-CN)
   - 英语 (en-US)

2. **翻译文件位置**
   - `/web/vue/src/locales/zh-CN.js` - 中文翻译
   - `/web/vue/src/locales/en-US.js` - 英文翻译
   - `/web/vue/src/locales/index.js` - i18n配置

3. **语言切换组件**
   - `/web/vue/src/components/common/LanguageSwitcher.vue`
   - 支持实时切换语言
   - 语言偏好保存在localStorage

4. **已国际化的模块**
   - 通用组件 (common)
   - 导航菜单 (nav)
   - 登录页面 (login)
   - 任务管理 (task)
   - 任务节点 (host)
   - 用户管理 (user)
   - 系统管理 (system)
   - 任务日志 (taskLog)
   - 双因素认证 (twoFactor)
   - 系统安装 (install)
   - 消息提示 (message)

### 使用方法

```vue
<script setup>
import { useI18n } from 'vue-i18n'
const { t, locale } = useI18n()
</script>

<template>
  <div>{{ t('common.save') }}</div>
  <div>{{ t('message.confirmDeleteTask', { name: taskName }) }}</div>
</template>
```

## 后端国际化 (Go + 自定义i18n包)

### 已实现功能

1. **i18n包位置**
   - `/internal/modules/i18n/i18n.go`

2. **语言支持**
   - 简体中文 (zh-CN)
   - 英语 (en-US)

3. **自动语言检测**
   - 基于HTTP请求头 `Accept-Language`
   - 通过中间件自动设置语言环境

4. **已国际化的模块**
   - 路由中间件 (routers.go)
   - 用户管理 (user.go)
   - 错误消息
   - API响应消息

### 使用方法

```go
import "github.com/gocronx-team/gocron/internal/modules/i18n"

func Handler(c *gin.Context) {
    locale := getLocale(c)
    message := i18n.T(locale, "save_success")
    // 使用翻译后的消息
}
```

## 翻译键列表

### 通用翻译键

- `app_not_installed` - 应用未安装
- `unauthorized` - 无权限访问
- `auth_failed` - 认证失败
- `form_validation_failed` - 表单验证失败
- `save_success` / `save_failed` - 保存成功/失败
- `update_success` / `update_failed` - 更新成功/失败
- `delete_success` / `delete_failed` - 删除成功/失败

### 用户相关

- `username_exists` - 用户名已存在
- `email_exists` - 邮箱已存在
- `password_required` - 请输入密码
- `password_mismatch` - 密码不匹配
- `old_password_error` - 旧密码错误
- `username_password_error` - 用户名或密码错误

### 2FA相关

- `2fa_code_required` - 需要2FA验证码
- `2fa_code_error` - 2FA验证码错误

## 添加新翻译

### 前端

1. 在 `/web/vue/src/locales/zh-CN.js` 添加中文翻译
2. 在 `/web/vue/src/locales/en-US.js` 添加英文翻译
3. 在组件中使用 `t('your.translation.key')`

### 后端

1. 在 `/internal/modules/i18n/i18n.go` 的 `messages` map中添加翻译
2. 在代码中使用 `i18n.T(locale, "your_translation_key")`

## 测试

### 前端测试

1. 启动开发服务器: `npm run dev`
2. 点击语言切换器切换语言
3. 验证所有页面文本正确显示

### 后端测试

1. 发送带有 `Accept-Language: zh-CN` 的请求
2. 发送带有 `Accept-Language: en-US` 的请求
3. 验证API响应消息使用正确的语言

## 注意事项

1. 所有用户可见的文本都应该使用翻译键
2. 翻译键应该语义化，便于理解
3. 支持参数插值，如 `t('message.confirmDeleteTask', { name: 'Task1' })`
4. 语言偏好会保存在localStorage，刷新页面后保持
5. 后端根据HTTP请求头自动检测语言

## 未来扩展

如需添加更多语言支持：

1. 创建新的语言文件，如 `ja-JP.js` (日语)
2. 在 `/web/vue/src/locales/index.js` 中注册新语言
3. 在 `LanguageSwitcher.vue` 中添加新语言选项
4. 在后端 `i18n.go` 中添加对应的翻译映射
