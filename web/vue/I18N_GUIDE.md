# 多语言国际化使用指南

## 概述

本项目已集成 Vue I18n 9，支持中英文切换。

## 已完成的功能

### 1. 基础配置
- ✅ 安装 vue-i18n@9
- ✅ 创建语言包文件 (zh-CN.js, en-US.js)
- ✅ 配置 i18n 实例
- ✅ 在 main.js 中注册

### 2. 语言切换组件
- ✅ 创建 LanguageSwitcher.vue 组件
- ✅ 集成到导航菜单
- ✅ 集成到登录页面
- ✅ 语言选择持久化到 localStorage

### 3. 已国际化的页面
- ✅ 登录页面 (login.vue)
- ✅ 导航菜单 (navMenu.vue)
- ✅ 任务编辑页面 (task/edit.vue)
- ✅ 全局确认对话框

## 使用方法

### 在组件中使用多语言

#### Composition API (推荐)

```vue
<template>
  <div>
    <h1>{{ t('common.confirm') }}</h1>
    <el-button>{{ t('common.save') }}</el-button>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()

// 切换语言
const switchLanguage = () => {
  locale.value = locale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
}
</script>
```

#### Options API

```vue
<template>
  <div>
    <h1>{{ t('common.confirm') }}</h1>
  </div>
</template>

<script>
export default {
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  }
}
</script>
```

### 添加新的翻译

1. 在 `src/locales/zh-CN.js` 中添加中文翻译：
```javascript
export default {
  myModule: {
    title: '我的标题',
    description: '我的描述'
  }
}
```

2. 在 `src/locales/en-US.js` 中添加英文翻译：
```javascript
export default {
  myModule: {
    title: 'My Title',
    description: 'My Description'
  }
}
```

3. 在组件中使用：
```vue
<template>
  <h1>{{ t('myModule.title') }}</h1>
  <p>{{ t('myModule.description') }}</p>
</template>
```

### 动态标签宽度

根据语言调整表单标签宽度：

```vue
<el-form :label-width="locale === 'zh-CN' ? '120px' : '180px'">
  <!-- 表单项 -->
</el-form>
```

## 语言包结构

```
src/locales/
├── index.js      # i18n 配置
├── zh-CN.js      # 中文语言包
└── en-US.js      # 英文语言包
```

### 当前语言包内容

#### common (通用)
- confirm, cancel, save, delete, edit, search, tip
- confirmOperation

#### nav (导航)
- taskManage, taskNode, userManage, systemManage
- logout, changePassword, twoFactor

#### login (登录)
- title, username, password, verifyCode, login
- 各种 placeholder 和验证消息

#### task (任务)
- 任务相关的所有字段和提示信息
- 包括：name, tag, type, command, timeout 等

## 待完成的页面

以下页面还需要国际化：

- [ ] 任务列表页面 (task/list.vue)
- [ ] 任务日志页面 (taskLog/list.vue)
- [ ] 主机管理页面 (host/list.vue, host/edit.vue)
- [ ] 用户管理页面 (user/list.vue, user/edit.vue)
- [ ] 系统设置页面 (system/*)
- [ ] 双因素认证页面 (user/twoFactor.vue)

## 最佳实践

1. **统一命名规范**
   - 使用小驼峰命名：`userName` 而不是 `user_name`
   - 使用有意义的键名：`usernamePlaceholder` 而不是 `ph1`

2. **分组管理**
   - 按功能模块分组：common, nav, login, task 等
   - 避免嵌套过深，最多 2-3 层

3. **Placeholder 和提示**
   - 为每个输入框的 placeholder 创建单独的键
   - 为验证消息创建单独的键

4. **动态内容**
   - 使用插值：`t('message', { name: 'John' })`
   - 在语言包中：`message: 'Hello {name}'`

5. **标签宽度**
   - 英文通常比中文长，需要更宽的标签
   - 使用响应式标签宽度

## 测试

1. 启动开发服务器：
```bash
npm run dev
```

2. 访问 http://localhost:8080

3. 测试语言切换：
   - 在导航栏点击语言切换器
   - 在登录页面点击语言切换器
   - 刷新页面，确认语言设置被保存

## 故障排除

### 翻译不显示
- 检查语言包中是否存在对应的键
- 检查是否正确导入 `useI18n`
- 检查控制台是否有错误信息

### 语言切换不生效
- 检查 localStorage 中的 `locale` 值
- 确认 `locale.value` 被正确设置
- 尝试刷新页面

### 表单验证消息未翻译
- 使用 `computed` 包装 `formRules`
- 确保在 `computed` 中使用 `t()` 函数

## 扩展语言

如果需要添加更多语言（如日语、韩语）：

1. 创建新的语言文件：`src/locales/ja-JP.js`
2. 在 `src/locales/index.js` 中导入并注册
3. 在 `LanguageSwitcher.vue` 中添加选项

```javascript
// src/locales/index.js
import jaJP from './ja-JP'

const i18n = createI18n({
  // ...
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS,
    'ja-JP': jaJP  // 新增
  }
})
```
