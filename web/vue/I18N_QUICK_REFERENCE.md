# 多语言快速参考

## 快速开始

### 1. 在新组件中使用

```vue
<template>
  <div>
    <h1>{{ t('common.title') }}</h1>
    <el-button>{{ t('common.save') }}</el-button>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
const { t, locale } = useI18n()
</script>
```

### 2. 添加语言切换器

```vue
<template>
  <LanguageSwitcher />
</template>

<script setup>
import LanguageSwitcher from '@/components/common/LanguageSwitcher.vue'
</script>
```

### 3. 动态表单标签宽度

```vue
<el-form :label-width="locale === 'zh-CN' ? '120px' : '180px'">
```

## 常用翻译键

### 通用 (common)
```javascript
t('common.confirm')      // 确定 / Confirm
t('common.cancel')       // 取消 / Cancel
t('common.save')         // 保存 / Save
t('common.delete')       // 删除 / Delete
t('common.edit')         // 编辑 / Edit
t('common.search')       // 搜索 / Search
```

### 导航 (nav)
```javascript
t('nav.taskManage')      // 任务管理 / Task Management
t('nav.taskNode')        // 任务节点 / Task Node
t('nav.userManage')      // 用户管理 / User Management
t('nav.systemManage')    // 系统管理 / System Management
```

### 登录 (login)
```javascript
t('login.title')         // 用户登录 / User Login
t('login.username')      // 用户名 / Username
t('login.password')      // 密码 / Password
t('login.login')         // 登录 / Login
```

### 任务 (task)
```javascript
t('task.name')           // 任务名称 / Task Name
t('task.tag')            // 标签 / Tag
t('task.type')           // 任务类型 / Task Type
t('task.command')        // 命令 / Command
t('task.timeout')        // 任务超时时间 / Task Timeout
```

## 添加新翻译的步骤

### 步骤 1: 编辑中文语言包
文件：`src/locales/zh-CN.js`

```javascript
export default {
  myFeature: {
    title: '我的功能',
    button: '点击我'
  }
}
```

### 步骤 2: 编辑英文语言包
文件：`src/locales/en-US.js`

```javascript
export default {
  myFeature: {
    title: 'My Feature',
    button: 'Click Me'
  }
}
```

### 步骤 3: 在组件中使用
```vue
<template>
  <h1>{{ t('myFeature.title') }}</h1>
  <el-button>{{ t('myFeature.button') }}</el-button>
</template>
```

## 表单验证消息

使用 `computed` 包装规则：

```vue
<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const formRules = computed(() => ({
  username: [
    { required: true, message: t('login.usernameRequired'), trigger: 'blur' }
  ],
  password: [
    { required: true, message: t('login.passwordRequired'), trigger: 'blur' }
  ]
}))
</script>
```

## 编程式切换语言

```javascript
import { useI18n } from 'vue-i18n'

const { locale } = useI18n()

// 切换到英文
locale.value = 'en-US'

// 切换到中文
locale.value = 'zh-CN'

// 保存到 localStorage
localStorage.setItem('locale', locale.value)
```

## 当前支持的语言

- `zh-CN` - 简体中文
- `en-US` - English

## 文件位置

```
src/
├── locales/
│   ├── index.js          # i18n 配置
│   ├── zh-CN.js          # 中文翻译
│   └── en-US.js          # 英文翻译
├── components/
│   └── common/
│       └── LanguageSwitcher.vue  # 语言切换组件
└── main.js               # 注册 i18n
```

## 注意事项

1. **始终同步更新两个语言文件**
2. **使用有意义的键名**
3. **英文标签通常需要更宽的空间**
4. **表单验证规则使用 computed**
5. **语言切换会自动保存到 localStorage**
